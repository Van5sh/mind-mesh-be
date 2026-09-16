package guards

import (
	"context"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

type ProjectGuard struct {
	repo *repository.ProjectRepository
}

func NewProjectGuard(repo *repository.ProjectRepository) *ProjectGuard {
	return &ProjectGuard{repo: repo}
}

func (g *ProjectGuard) EnsureProjectExists(ctx context.Context, projectID pgtype.UUID) (database.Project, error) {
	project, err := g.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		if isNoRows(err) {
			return database.Project{}, apperrors.NotFoundError("project not found")
		}
		return database.Project{}, apperrors.InternalError("failed to fetch project", err)
	}
	return project, nil
}

func (g *ProjectGuard) EnsureProjectMember(ctx context.Context, projectID, userID pgtype.UUID) (database.ProjectMember, error) {
	member, err := g.repo.GetProjectMember(ctx, database.GetProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		if isNoRows(err) {
			return database.ProjectMember{}, apperrors.ForbiddenError("user is not a project member")
		}
		return database.ProjectMember{}, apperrors.InternalError("failed to fetch project member", err)
	}
	return member, nil
}

func (g *ProjectGuard) EnsureProjectOwner(ctx context.Context, projectID, userID pgtype.UUID) (database.Project, error) {
	project, err := g.EnsureProjectExists(ctx, projectID)
	if err != nil {
		return database.Project{}, err
	}
	if !sameUUID(project.OwnerID, userID) {
		return database.Project{}, apperrors.ForbiddenError("user is not the project owner")
	}
	return project, nil
}

func (g *ProjectGuard) EnsureProjectNameAvailableForOwner(ctx context.Context, ownerID pgtype.UUID, name string) error {
	_, err := g.repo.GetProjectByName(ctx, database.GetProjectByNameParams{
		Name:    name,
		OwnerID: ownerID,
	})
	if err == nil {
		return apperrors.ConflictError("project name already exists")
	}
	if isNoRows(err) {
		return nil
	}
	return apperrors.InternalError("failed to check project name", err)
}
