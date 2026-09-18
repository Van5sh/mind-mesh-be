package graph

import (
	"context"
	"errors"
	"fmt"

	"example/hello/graph/loaders"
	"example/hello/graph/model"
	"example/hello/internal/database"
)

// loadUser fetches the full user row behind a *model.User.
//
// Many resolvers only know a user's ID and hand back a stub
// (&model.User{ID: ...}); the User field resolvers call this to fill in the
// rest. Inside a GraphQL operation the lookup goes through the per-operation
// dataloader, so N stubs resolve with one GetUsersByIDs query. Without a
// loader on the context (e.g. a resolver called directly from a test) it falls
// back to a single GetUserByID.
func (r *Resolver) loadUser(ctx context.Context, obj *model.User) (database.User, error) {
	id, err := parseUUID(obj.ID)
	if err != nil {
		return database.User{}, err
	}

	if l := loaders.From(ctx); l != nil {
		user, err := l.UserByID.Load(ctx, id)
		if err != nil {
			if errors.Is(err, loaders.ErrNotFound) {
				return database.User{}, fmt.Errorf("user %s not found", obj.ID)
			}
			return database.User{}, err
		}
		return user, nil
	}

	return r.App.Services.User.GetUserByID(ctx, id)
}
