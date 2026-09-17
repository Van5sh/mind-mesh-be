package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"example/hello/graph"
	graphresolver "example/hello/graph/resolver"
	"example/hello/internal/app"
	"example/hello/internal/auth"
	"example/hello/internal/services/aws"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/coder/websocket"
	"github.com/gofiber/fiber"
	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8090"

// defaultWSPort is where GraphQL subscriptions (WebSocket) are served -
// deliberately a *separate* listener from defaultPort. See
// BACKEND_HANDOFF.md's "Real-time chat" section for why: this server uses
// Fiber v1 over fasthttp for the main port, and fasthttp's net/http bridge
// (fasthttpadaptor) can't perform a WebSocket upgrade - it buffers the
// whole response instead of exposing a hijackable connection. Subscriptions
// are instead served by a second, plain net/http listener running the same
// gqlgen handler (same schema, resolvers, and auth middleware - just a
// second transport door into the identical server).
const defaultWSPort = "8081"

func StartServer() {

	// ============================================================
	// Environment
	// ============================================================

	if err := godotenv.Load(); err != nil {
		log.Println(
			"Warning: .env file not found, using system environment variables",
		)
	}

	ctx := context.Background()

	awsConfig, err := aws.InitializeAWSConfig(ctx)
	if err != nil {
		log.Fatalf(
			"initialize AWS configuration: %v",
			err,
		)
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = defaultPort
	}

	application, err := app.New(
		ctx,
		os.Getenv("DATABASE_URL"),
		awsConfig,
	)

	if err != nil {
		log.Fatalf(
			"initialize application: %v",
			err,
		)
	}

	defer application.Close()

	log.Println(
		"Application initialized successfully",
	)

	oauthHandler, err :=
		auth.NewOAuthHandlerFromEnvironment(
			ctx,
			application.Repositories.User,
			application.Repositories.OAuth,
			application.Services.Session,
		)

	if err != nil {
		log.Fatalf(
			"initialize authentication: %v",
			err,
		)
	}
	server := fiber.New()

	srv := handler.New(
		graph.NewExecutableSchema(
			graph.Config{
				Resolvers: graphresolver.NewResolver(
					application,
				),
			},
		),
	)

	srv.AddTransport(
		transport.Options{},
	)

	srv.AddTransport(
		transport.GET{},
	)

	srv.AddTransport(
		transport.POST{},
	)

	srv.AddTransport(
		transport.MultipartForm{},
	)

	srv.AddTransport(
		transport.Websocket{
			KeepAlivePingInterval: 15 * time.Second,
			Implementation: transport.CoderWebsocketImplementation{
				AcceptOptions: websocket.AcceptOptions{
					// The API and frontend are different origins in every
					// real deployment, so the default (same-origin only)
					// would reject every browser subscription. Restrict
					// this to exactly the configured frontend, matching
					// the HTTP CORS policy above.
					OriginPatterns: []string{
						oauthHandler.FrontendOrigin(),
					},
				},
			},
		},
	)

	srv.SetQueryCache(
		lru.New[*ast.QueryDocument](1000),
	)

	srv.Use(
		extension.Introspection{},
	)

	srv.Use(
		extension.AutomaticPersistedQuery{
			Cache: lru.New[string](100),
		},
	)

	graphqlHandler := oauthHandler.Middleware(
		srv,
	)

	graphqlHandler = oauthHandler.CORS(
		graphqlHandler,
	)

	playgroundHandler := playground.Handler(
		"GraphQL Playground",
		"/query",
	)

	server.Get(
		"/",
		func(c *fiber.Ctx) {

			fasthttpadaptor.NewFastHTTPHandler(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						playgroundHandler.ServeHTTP(
							w,
							r,
						)
					},
				),
			)(c.Fasthttp)
		},
	)

	server.All(
		"/query",
		func(c *fiber.Ctx) {

			fasthttpadaptor.NewFastHTTPHandler(
				graphqlHandler,
			)(c.Fasthttp)
		},
	)

	// The actual Google/GitHub OAuth handshake happens client-side via the
	// Firebase JS SDK now - this is the only auth route left on this
	// server, and it just verifies the Firebase ID token the frontend
	// already has afterward. See BACKEND_HANDOFF.md §3.
	server.Post(
		"/auth/firebase",
		func(c *fiber.Ctx) {

			fasthttpadaptor.NewFastHTTPHandler(
				http.HandlerFunc(
					oauthHandler.FirebaseLogin,
				),
			)(c.Fasthttp)
		},
	)

	server.Post(
		"/auth/logout",
		func(c *fiber.Ctx) {

			fasthttpadaptor.NewFastHTTPHandler(
				http.HandlerFunc(
					oauthHandler.Logout,
				),
			)(c.Fasthttp)
		},
	)

	log.Printf(
		"🚀 Server ready at http://localhost:%s/",
		port,
	)

	log.Printf(
		"🔍 GraphQL Playground at http://localhost:%s/",
		port,
	)

	log.Printf(
		"📡 GraphQL endpoint at http://localhost:%s/query",
		port,
	)

	log.Printf(
		"☁️  AWS ready - Region: %s, S3 Bucket: %s, DynamoDB Table: %s",
		awsConfig.Region,
		awsConfig.S3Bucket,
		awsConfig.DynamoDBTable,
	)

	wsPort := os.Getenv("WS_PORT")
	if wsPort == "" {
		wsPort = defaultWSPort
	}

	// Runs the identical graphqlHandler (same schema, resolvers, auth) on
	// a second, plain net/http listener so GraphQL subscriptions
	// (WebSocket) work - see the defaultWSPort comment above for why this
	// can't just be another route on the Fiber server above.
	go func() {
		log.Printf(
			"📡 GraphQL subscriptions (WebSocket) at ws://localhost:%s/query",
			wsPort,
		)

		if err := http.ListenAndServe(":"+wsPort, graphqlHandler); err != nil {
			log.Fatalf(
				"start websocket server: %v",
				err,
			)
		}
	}()

	if err := server.Listen(":" + port); err != nil {
		log.Fatalf(
			"start server: %v",
			err,
		)
	}
}

func main() {
	StartServer()
}
