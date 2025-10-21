package main

import (
	"log"
	"net/http"
	"os"

	"example/hello/graph"
	"example/hello/graph/postgres"

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
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	postgres.InitDB()
	log.Println("Database initialized successfully")
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	app := fiber.New()

	resolver := &graph.Resolver{
		UsersData: postgres.UsersRepo{},
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	playgroundHandler := playground.Handler("GraphQL Playground", "/query")

	app.Get("/", func(c *fiber.Ctx) {
		fasthttpadaptor.NewFastHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			playgroundHandler.ServeHTTP(w, r)
		}))(c.Fasthttp)
	})

	app.Post("/query", func(c *fiber.Ctx) {
		fasthttpadaptor.NewFastHTTPHandler(srv)(c.Fasthttp)
	})

	app.Get("/query", func(c *fiber.Ctx) {
		fasthttpadaptor.NewFastHTTPHandler(srv)(c.Fasthttp)
	})

	log.Printf("🚀 Server ready at http://localhost:%s/", port)
	log.Printf("🔍 GraphQL Playground at http://localhost:%s/", port)
	log.Printf("📡 GraphQL endpoint at http://localhost:%s/query", port)

	app.Listen(port)
}

func main() {
	StartServer()
}
