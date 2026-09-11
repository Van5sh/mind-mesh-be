package graph

import "example/hello/internal/app"

// Resolver contains all application dependencies required
// by GraphQL resolvers.
type Resolver struct {
	App *app.App
}

// NewResolver creates a GraphQL resolver dependency container.
func NewResolver(application *app.App) *Resolver {
	return &Resolver{
		App: application,
	}
}
