package graph

import (
	"context"
	"errors"
	"fmt"

	"example/hello/graph/loaders"
	"example/hello/graph/model"
	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

// isStubProject reports whether obj is an ID-only placeholder
// (&model.Project{ID: ...}) rather than a fully loaded project. Project names
// are validated as required, so a real project never has an empty one.
func isStubProject(obj *model.Project) bool {
	return obj.Name == ""
}

// loadProject fetches the full project row behind a stub *model.Project.
//
// Helpers for files, folders, chats, reports, flowcharts, activity logs and
// members only know a project's ID and hand back a stub; the Project field
// resolvers call this to fill in the rest. Inside a GraphQL operation the
// lookup goes through the per-operation dataloader, so N stubs resolve with
// one GetProjectsByIDs query. Without a loader on the context it falls back
// to a single GetProjectByID.
func (r *Resolver) loadProject(ctx context.Context, obj *model.Project) (database.Project, error) {
	id, err := parseUUID(obj.ID)
	if err != nil {
		return database.Project{}, err
	}

	if l := loaders.From(ctx); l != nil {
		project, err := l.ProjectByID.Load(ctx, id)
		if err != nil {
			if errors.Is(err, loaders.ErrNotFound) {
				return database.Project{}, fmt.Errorf("project %s not found", obj.ID)
			}
			return database.Project{}, err
		}
		return project, nil
	}

	return r.App.Services.Project.GetProjectByID(ctx, id)
}

// loadFolder fetches one folder row through the per-operation dataloader
// (batched across every folder resolved in the operation), or directly when no
// loader is installed.
func (r *Resolver) loadFolder(ctx context.Context, id pgtype.UUID) (database.Folder, error) {
	if l := loaders.From(ctx); l != nil {
		folder, err := l.FolderByID.Load(ctx, id)
		if errors.Is(err, loaders.ErrNotFound) {
			return database.Folder{}, fmt.Errorf("folder not found")
		}
		return folder, err
	}

	return r.App.Services.File.GetFolderByID(ctx, id)
}
