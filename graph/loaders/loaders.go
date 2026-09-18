// Package loaders holds per-operation dataloaders that batch the "one query
// per parent" lookups GraphQL nested fields would otherwise cause (the N+1
// problem): resolving Project.files for 20 projects used to run 20 SELECTs;
// with a loader every Load call made while resolving one operation is
// collected and fetched in a single query.
package loaders

import (
	"context"
	"time"

	"example/hello/internal/database"

	"github.com/99designs/gqlgen/graphql"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vikstrous/dataloadgen"
)

// FilesFetcher is the batched lookup the files loader is built on.
// services.FileService satisfies it.
type FilesFetcher interface {
	GetFilesByProjectIDs(ctx context.Context, projectIDs []pgtype.UUID) (map[pgtype.UUID][]database.File, error)
}

// UsersFetcher is the batched user lookup the user loader is built on.
// services.UserService satisfies it.
type UsersFetcher interface {
	GetUsersByIDs(ctx context.Context, ids []pgtype.UUID) ([]database.User, error)
}

// ErrNotFound is what UserByID.Load returns for an ID that has no user.
var ErrNotFound = dataloadgen.ErrNotFound

type Loaders struct {
	FilesByProject *dataloadgen.Loader[pgtype.UUID, []database.File]
	UserByID       *dataloadgen.Loader[pgtype.UUID, database.User]
}

// batchWait is how long the loader waits for more keys before firing a
// batch. Resolvers for sibling objects run back to back, so a short window
// is enough to collect them all without adding noticeable latency.
const batchWait = 2 * time.Millisecond

func New(files FilesFetcher, users UsersFetcher) *Loaders {
	return &Loaders{
		FilesByProject: dataloadgen.NewLoader(
			func(ctx context.Context, projectIDs []pgtype.UUID) ([][]database.File, []error) {
				grouped, err := files.GetFilesByProjectIDs(ctx, projectIDs)
				if err != nil {
					errs := make([]error, len(projectIDs))
					for i := range errs {
						errs[i] = err
					}
					return nil, errs
				}

				// Results must line up with the keys, in order.
				out := make([][]database.File, len(projectIDs))
				for i, id := range projectIDs {
					out[i] = grouped[id]
				}
				return out, nil
			},
			dataloadgen.WithWait(batchWait),
		),

		// A mapped loader, not a positional one: GetUsersByIDs returns rows
		// ordered by username (not by the requested IDs) and simply omits
		// IDs that don't exist. Matching by key makes the order irrelevant
		// and gives every missing ID ErrNotFound instead of a wrong user.
		UserByID: dataloadgen.NewMappedLoader(
			func(ctx context.Context, ids []pgtype.UUID) (map[pgtype.UUID]database.User, error) {
				found, err := users.GetUsersByIDs(ctx, ids)
				if err != nil {
					return nil, err
				}

				byID := make(map[pgtype.UUID]database.User, len(found))
				for _, u := range found {
					byID[u.ID] = u
				}
				return byID, nil
			},
			dataloadgen.WithWait(batchWait),
		),
	}
}

type ctxKey struct{}

// From returns the loaders for the current operation, or nil when none were
// installed (e.g. a resolver called directly from a test).
func From(ctx context.Context) *Loaders {
	l, _ := ctx.Value(ctxKey{}).(*Loaders)
	return l
}

// OperationMiddleware installs a fresh set of loaders for every GraphQL
// operation (register it with srv.AroundOperations). Per operation rather
// than per HTTP request, on purpose: a WebSocket connection is one long-lived
// request carrying many operations, and a loader's cache must not outlive the
// operation it was created for or later queries would see stale data.
func OperationMiddleware(files FilesFetcher, users UsersFetcher) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		return next(context.WithValue(ctx, ctxKey{}, New(files, users)))
	}
}
