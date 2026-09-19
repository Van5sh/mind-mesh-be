// Package loaders holds per-operation dataloaders that batch the "one query
// per parent" lookups GraphQL nested fields would otherwise cause (the N+1
// problem): resolving Project.files (or members, chats, ...) for 20 projects
// used to run 20 SELECTs; with a loader every Load call made while resolving
// one operation is collected and fetched in a single query.
//
// There are two shapes of loader, each built by one generic constructor:
//
//   - list loaders (byKey): "the X of this parent" - a slice per key, where a
//     parent with nothing loads as an empty list (Project.files, Chat.messages,
//     User.ownedProjects, ...);
//   - lookup loaders (byID): "the one row for this ID" - a single value per
//     key, where a missing row loads as ErrNotFound (User by ID, File AI
//     metadata, ...).
package loaders

import (
	"context"
	"time"

	"example/hello/internal/database"

	"github.com/99designs/gqlgen/graphql"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vikstrous/dataloadgen"
)

// Grouped is a batch function that fetches the rows of many parents in one
// query and returns them grouped by parent ID. Every requested ID must be
// present in the map (nil slice for a parent with no rows) - the services'
// GetXByYIDs methods already guarantee that.
type Grouped[T any] func(ctx context.Context, ids []pgtype.UUID) (map[pgtype.UUID][]T, error)

// Rows is a batch function that fetches one row per ID in a single query and
// returns them flat, in any order; IDs with no row are simply absent.
type Rows[T any] func(ctx context.Context, ids []pgtype.UUID) ([]T, error)

// Sources bundles every batch function the loaders read from - one field per
// loader, each satisfied by a service method (see cmd/server/main.go). Leave a
// field nil to skip that loader (its Load must then not be called) - handy in
// tests.
type Sources struct {
	// Lists per project.
	FilesByProject      Grouped[database.File]
	FoldersByProject    Grouped[database.Folder]
	MembersByProject    Grouped[database.ProjectMember]
	ChatsByProject      Grouped[database.Chat]
	ReportsByProject    Grouped[database.GetReportsByProjectIDsRow]
	FlowchartsByProject Grouped[database.Flowchart]
	ActivityByProject   Grouped[database.ActivityLog]

	// Lists per chat / file / folder / user.
	MessagesByChat       Grouped[database.ChatMessage]
	ParticipantsByChat   Grouped[database.ChatParticipant]
	SharesByFile         Grouped[database.FileShare]
	FilesByFolder        Grouped[database.File]
	ChildFoldersByParent Grouped[database.Folder]
	OwnedProjectsByOwner Grouped[database.Project]
	MembershipsByUser    Grouped[database.ProjectMember]
	SharesBySharedWith   Grouped[database.FileShare]

	// One row per ID.
	Users      Rows[database.User]
	Projects   Rows[database.Project]
	Folders    Rows[database.Folder]
	Profiles   Rows[database.UserProfile]
	AIMetadata Rows[database.FileAiMetadatum]
	Storages   Rows[database.FileStorage]
}

// ErrNotFound is what a lookup loader's Load returns for an ID that has no
// row.
var ErrNotFound = dataloadgen.ErrNotFound

type (
	// ListLoader loads "the rows of this parent".
	ListLoader[T any] = dataloadgen.Loader[pgtype.UUID, []T]
	// LookupLoader loads "the row with this ID".
	LookupLoader[T any] = dataloadgen.Loader[pgtype.UUID, T]
)

type Loaders struct {
	FilesByProject      *ListLoader[database.File]
	FoldersByProject    *ListLoader[database.Folder]
	MembersByProject    *ListLoader[database.ProjectMember]
	ChatsByProject      *ListLoader[database.Chat]
	ReportsByProject    *ListLoader[database.GetReportsByProjectIDsRow]
	FlowchartsByProject *ListLoader[database.Flowchart]
	ActivityByProject   *ListLoader[database.ActivityLog]

	MessagesByChat       *ListLoader[database.ChatMessage]
	ParticipantsByChat   *ListLoader[database.ChatParticipant]
	SharesByFile         *ListLoader[database.FileShare]
	FilesByFolder        *ListLoader[database.File]
	ChildFoldersByParent *ListLoader[database.Folder]
	OwnedProjectsByOwner *ListLoader[database.Project]
	MembershipsByUser    *ListLoader[database.ProjectMember]
	SharesBySharedWith   *ListLoader[database.FileShare]

	UserByID         *LookupLoader[database.User]
	ProjectByID      *LookupLoader[database.Project]
	FolderByID       *LookupLoader[database.Folder]
	ProfileByUser    *LookupLoader[database.UserProfile]
	AIMetadataByFile *LookupLoader[database.FileAiMetadatum]
	StorageByFile    *LookupLoader[database.FileStorage]
}

// batchWait is how long a loader waits for more keys before firing a batch.
// Resolvers for sibling objects run back to back, so a short window is enough
// to collect them all without adding noticeable latency.
const batchWait = 2 * time.Millisecond

// byKey builds a list loader from a Grouped batch function. The function is
// looked up lazily (at fetch time, not when the loader is built) so an unset
// source only fails if its loader is actually used.
func byKey[T any](fetch func() Grouped[T]) *ListLoader[T] {
	return dataloadgen.NewLoader(
		func(ctx context.Context, ids []pgtype.UUID) ([][]T, []error) {
			grouped, err := fetch()(ctx, ids)
			if err != nil {
				// One error per key, and - easy to miss - still one (empty)
				// value per key: dataloadgen rejects a short values slice
				// with a "bug in fetch function" error that would mask the
				// real failure.
				errs := make([]error, len(ids))
				for i := range errs {
					errs[i] = err
				}
				return make([][]T, len(ids)), errs
			}

			// Results must line up with the keys, in order.
			out := make([][]T, len(ids))
			for i, id := range ids {
				out[i] = grouped[id]
			}
			return out, nil
		},
		dataloadgen.WithWait(batchWait),
	)
}

// byID builds a lookup loader from a Rows batch function and a way to read a
// row's own ID. A mapped loader, not a positional one: the batch query returns
// rows in whatever order SQL likes and omits IDs that don't exist, so matching
// by key makes the order irrelevant and gives every missing ID ErrNotFound
// instead of somebody else's row.
func byID[T any](fetch func() Rows[T], idOf func(T) pgtype.UUID) *LookupLoader[T] {
	return dataloadgen.NewMappedLoader(
		func(ctx context.Context, ids []pgtype.UUID) (map[pgtype.UUID]T, error) {
			rows, err := fetch()(ctx, ids)
			if err != nil {
				return nil, err
			}

			byID := make(map[pgtype.UUID]T, len(rows))
			for _, row := range rows {
				byID[idOf(row)] = row
			}
			return byID, nil
		},
		dataloadgen.WithWait(batchWait),
	)
}

// New returns a fresh set of loaders reading from src.
func New(src Sources) *Loaders {
	return &Loaders{
		FilesByProject:      byKey(func() Grouped[database.File] { return src.FilesByProject }),
		FoldersByProject:    byKey(func() Grouped[database.Folder] { return src.FoldersByProject }),
		MembersByProject:    byKey(func() Grouped[database.ProjectMember] { return src.MembersByProject }),
		ChatsByProject:      byKey(func() Grouped[database.Chat] { return src.ChatsByProject }),
		ReportsByProject:    byKey(func() Grouped[database.GetReportsByProjectIDsRow] { return src.ReportsByProject }),
		FlowchartsByProject: byKey(func() Grouped[database.Flowchart] { return src.FlowchartsByProject }),
		ActivityByProject:   byKey(func() Grouped[database.ActivityLog] { return src.ActivityByProject }),

		MessagesByChat:       byKey(func() Grouped[database.ChatMessage] { return src.MessagesByChat }),
		ParticipantsByChat:   byKey(func() Grouped[database.ChatParticipant] { return src.ParticipantsByChat }),
		SharesByFile:         byKey(func() Grouped[database.FileShare] { return src.SharesByFile }),
		FilesByFolder:        byKey(func() Grouped[database.File] { return src.FilesByFolder }),
		ChildFoldersByParent: byKey(func() Grouped[database.Folder] { return src.ChildFoldersByParent }),
		OwnedProjectsByOwner: byKey(func() Grouped[database.Project] { return src.OwnedProjectsByOwner }),
		MembershipsByUser:    byKey(func() Grouped[database.ProjectMember] { return src.MembershipsByUser }),
		SharesBySharedWith:   byKey(func() Grouped[database.FileShare] { return src.SharesBySharedWith }),

		UserByID: byID(
			func() Rows[database.User] { return src.Users },
			func(u database.User) pgtype.UUID { return u.ID },
		),
		ProjectByID: byID(
			func() Rows[database.Project] { return src.Projects },
			func(p database.Project) pgtype.UUID { return p.ID },
		),
		FolderByID: byID(
			func() Rows[database.Folder] { return src.Folders },
			func(f database.Folder) pgtype.UUID { return f.ID },
		),
		ProfileByUser: byID(
			func() Rows[database.UserProfile] { return src.Profiles },
			func(p database.UserProfile) pgtype.UUID { return p.UserID },
		),
		AIMetadataByFile: byID(
			func() Rows[database.FileAiMetadatum] { return src.AIMetadata },
			func(m database.FileAiMetadatum) pgtype.UUID { return m.FileID },
		),
		StorageByFile: byID(
			func() Rows[database.FileStorage] { return src.Storages },
			func(s database.FileStorage) pgtype.UUID { return s.FileID },
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
func OperationMiddleware(src Sources) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		return next(context.WithValue(ctx, ctxKey{}, New(src)))
	}
}
