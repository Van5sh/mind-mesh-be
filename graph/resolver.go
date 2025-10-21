package graph

import (
	"example/hello/graph/postgres"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	UsersData postgres.UsersRepo
}
