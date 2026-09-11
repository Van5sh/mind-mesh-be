package repository

import (
	"context"
	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectRepository struct {
	db *pgxpool.Pool
	q  *database.Queries
}

func NewProjectRepository(db *pgxpool.Pool, q *database.Queries) *ProjectRepository {
	return &ProjectRepository{
		db: db,
		q:  q,
	}
}

func (r *ProjectRepository) AddProjectMember(ctx context.Context, params database.AddProjectMemberParams) (database.ProjectMember, error) {
	return r.q.AddProjectMember(ctx, params)
}

func (r *ProjectRepository) ArchiveProject(ctx context.Context, id pgtype.UUID) error {
	return r.q.ArchiveProject(ctx, id)
}

func (r *ProjectRepository) CreateProject(ctx context.Context, params database.CreateProjectParams) (database.Project, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return database.Project{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := r.q.WithTx(tx)
	project, err := queries.CreateProject(ctx, params)
	if err != nil {
		return database.Project{}, err
	}
	if _, err := queries.AddProjectMember(ctx, database.AddProjectMemberParams{
		ProjectID: project.ID,
		UserID:    project.OwnerID,
		Role:      database.ProjectRoleOWNER,
	}); err != nil {
		return database.Project{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return database.Project{}, err
	}
	return project, nil
}

func (r *ProjectRepository) DeleteProject(ctx context.Context, id pgtype.UUID) error {
	return r.q.DeleteProject(ctx, id)
}

func (r *ProjectRepository) GetArchivedProjectsByOwner(ctx context.Context, ownerID pgtype.UUID) ([]database.Project, error) {
	return r.q.GetArchivedProjectsByOwner(ctx, ownerID)
}

func (r *ProjectRepository) GetArchivedProjectsForUser(ctx context.Context, ownerID pgtype.UUID) ([]database.Project, error) {
	return r.q.GetArchivedProjectsForUser(ctx, ownerID)
}

func (r *ProjectRepository) GetProjectByID(ctx context.Context, id pgtype.UUID) (database.Project, error) {
	return r.q.GetProjectByID(ctx, id)
}

func (r *ProjectRepository) GetProjectByName(ctx context.Context, params database.GetProjectByNameParams) (database.Project, error) {
	return r.q.GetProjectByName(ctx, params)
}

func (r *ProjectRepository) GetProjectMember(ctx context.Context, params database.GetProjectMemberParams) (database.ProjectMember, error) {
	return r.q.GetProjectMember(ctx, params)
}

func (r *ProjectRepository) GetProjectMembers(ctx context.Context, projectID pgtype.UUID) ([]database.ProjectMember, error) {
	return r.q.GetProjectMembers(ctx, projectID)
}

func (r *ProjectRepository) GetProjectsByOwnerID(ctx context.Context, ownerID pgtype.UUID) ([]database.Project, error) {
	return r.q.GetProjectsByOwnerID(ctx, ownerID)
}

func (r *ProjectRepository) GetProjectsForUser(ctx context.Context, ownerID pgtype.UUID) ([]database.Project, error) {
	return r.q.GetProjectsForUser(ctx, ownerID)
}

func (r *ProjectRepository) RemoveProjectMember(ctx context.Context, params database.RemoveProjectMemberParams) error {
	return r.q.RemoveProjectMember(ctx, params)
}

func (r *ProjectRepository) RestoreProject(ctx context.Context, id pgtype.UUID) error {
	return r.q.RestoreProject(ctx, id)
}

func (r *ProjectRepository) TransferOwnership(ctx context.Context, params database.TransferOwnershipParams) (database.Project, error) {
	return r.q.TransferOwnership(ctx, params)
}

func (r *ProjectRepository) UpdateProject(ctx context.Context, params database.UpdateProjectParams) (database.Project, error) {
	return r.q.UpdateProject(ctx, params)
}

func (r *ProjectRepository) UpdateProjectMemberRole(ctx context.Context, params database.UpdateProjectMemberRoleParams) (database.ProjectMember, error) {
	return r.q.UpdateProjectMemberRole(ctx, params)
}
