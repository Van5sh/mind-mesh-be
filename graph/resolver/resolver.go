package graph

import "example/hello/internal/app"

// Resolver serves as the dependency-injection container
// for all GraphQL resolvers.
type Resolver struct {
	App *app.App
}

func NewResolver(app *app.App) *Resolver {
	return &Resolver{
		App: app,
	}
}
