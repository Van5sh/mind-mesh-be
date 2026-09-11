package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"example/hello/graph"
	graphresolver "example/hello/graph/resolver"
	"example/hello/internal/app"
	"example/hello/internal/auth"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/fiber"
	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

func StartServer() {

	// ============================================================
	// Environment
	// ============================================================

	if err := godotenv.Load(); err != nil {
		log.Println(
			"Warning: .env file not found, using system environment variables",
		)
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = defaultPort
	}

	application, err := app.New(
		context.Background(),
		os.Getenv("DATABASE_URL"),
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
			application.Repositories.User,
			application.Repositories.OAuth,
			application.Services.Session,
		)

	if err != nil {
		log.Fatalf(
			"initialize OAuth: %v",
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

	server.Get(
		"/auth/google",
		func(c *fiber.Ctx) {

			fasthttpadaptor.NewFastHTTPHandler(
				http.HandlerFunc(
					oauthHandler.GoogleLogin,
				),
			)(c.Fasthttp)
		},
	)

	server.Get(
		"/auth/google/callback",
		func(c *fiber.Ctx) {

			fasthttpadaptor.NewFastHTTPHandler(
				http.HandlerFunc(
					oauthHandler.GoogleCallback,
				),
			)(c.Fasthttp)
		},
	)

	server.Get(
		"/auth/github",
		func(c *fiber.Ctx) {

			fasthttpadaptor.NewFastHTTPHandler(
				http.HandlerFunc(
					oauthHandler.GitHubLogin,
				),
			)(c.Fasthttp)
		},
	)

	server.Get(
		"/auth/github/callback",
		func(c *fiber.Ctx) {

			fasthttpadaptor.NewFastHTTPHandler(
				http.HandlerFunc(
					oauthHandler.GitHubCallback,
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
