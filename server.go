package main

import (
	"context"
	"example/hello/graph"
	graphresolver "example/hello/graph/resolver"
	"example/hello/internal/app"
	"example/hello/internal/services/aws"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

func main() {
	ctx := context.Background()
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize AWS configuration
	awsConfig, err := aws.InitializeAWSConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize AWS configuration: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	application, err := app.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("initialize application: %v", err)
	}
	defer application.Close()

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: graphresolver.NewResolver(application)}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("AWS configuration initialized successfully")
	log.Printf("S3 Bucket: %s", awsConfig.S3Bucket)
	log.Printf("DynamoDB Table: %s", awsConfig.DynamoDBTable)
	log.Printf("SES From Email: %s", awsConfig.SESFromEmail)
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
