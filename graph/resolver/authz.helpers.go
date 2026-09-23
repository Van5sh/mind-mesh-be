package graph

import (
	"context"

	"example/hello/internal/apperrors"
	"example/hello/internal/auth"
	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

// currentUserID returns the authenticated user's ID, or an Unauthorized
// error if the request has no valid session.
func currentUserID(ctx context.Context) (pgtype.UUID, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return pgtype.UUID{}, apperrors.UnauthorizedError("authentication required")
	}
	return userID, nil
}

// requireProjectMember authorizes the authenticated user against a
// project: the owner or any explicit member passes; anyone else gets a
// NotFound (project doesn't exist) or Forbidden (not a member) error.
// Resolvers for project-scoped resources call this before reading or
// mutating them.
func (r *Resolver) requireProjectMember(ctx context.Context, projectID pgtype.UUID) error {
	userID, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	return r.App.Services.Project.EnsureMemberAccess(ctx, projectID, userID)
}

// requireProjectOwner authorizes only the project's owner - for
// destructive or ownership-sensitive operations (delete, archive,
// transfer ownership, membership management).
func (r *Resolver) requireProjectOwner(ctx context.Context, projectID pgtype.UUID) error {
	userID, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	return r.App.Services.Project.EnsureOwnerAccess(ctx, projectID, userID)
}

// requireFileAccess authorizes the authenticated user against a file that
// may or may not belong to a project. A project file uses project
// membership, unchanged. A personal file (no project) is visible only to
// whoever uploaded it, or to someone it's been explicitly shared with via
// ShareFile - both existing mechanisms, not a new permission system.
func (r *Resolver) requireFileAccess(ctx context.Context, file database.File) error {
	if file.ProjectID.Valid {
		return r.requireProjectMember(ctx, file.ProjectID)
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	if file.UploadedBy.Valid && file.UploadedBy == userID {
		return nil
	}

	if _, err := r.App.Services.File.GetFileShareByUser(ctx, file.ID, userID); err != nil {
		return apperrors.ForbiddenError("you do not have access to this file")
	}
	return nil
}
