package repository

import (
	"context"
	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

type ProjectRepository struct {
	q *database.Queries
}

func NewProjectRepository(q *database.Queries) *ProjectRepository {
	return &ProjectRepository{
		q: q,
	}
}

func (r *ProjectRepository) AddProjectMember(ctx context.Context, params database.AddProjectMemberParams) (database.ProjectMember, error) {
	return r.q.AddProjectMember(ctx, params)
}

func (r *ProjectRepository) ArchiveProject(ctx context.Context, id pgtype.UUID) error {
	return r.q.ArchiveProject(ctx, id)
}

func (r *ProjectRepository) CreateProject(ctx context.Context, params database.CreateProjectParams) (database.Project, error) {
	return r.q.CreateProject(ctx, params)
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
