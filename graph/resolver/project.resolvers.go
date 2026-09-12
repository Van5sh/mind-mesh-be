package graph

import (
	"context"
	"example/hello/graph/helpers"
	"example/hello/graph/model"
	"example/hello/internal/apperrors"
	"example/hello/internal/auth"
	"example/hello/internal/database"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

// CreateProject is the resolver for the createProject field.
func (r *mutationResolver) CreateProject(ctx context.Context, input model.CreateProjectInput) (*model.Project, error) {
	userId, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, apperrors.UnauthorizedError(
			"authentication required",
		)
	}
	var description pgtype.Text

	if input.Description != nil {
		description = pgtype.Text{
			String: *input.Description,
			Valid:  true,
		}
	}

	project, err := r.App.Services.Project.CreateProject(
		ctx,
		database.CreateProjectParams{
			OwnerID:     userId,
			Name:        input.Name,
			Description: description,
		},
	)
	if err != nil {
		return nil, err
	}

	result := helpers.ProjectToModel(project)

	// Load owner because Project.Owner is required by GraphQL.
	user, err := r.App.Services.User.GetUserByID(ctx, userId)
	if err != nil {
		return nil, err
	}

	result.Owner = &model.User{
		ID:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	}

	return result, nil
}

// UpdateProject is the resolver for the updateProject field.
func (r *mutationResolver) UpdateProject(ctx context.Context, id string, input model.UpdateProjectInput) (*model.Project, error) {
	projectID, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	// If your GraphQL schema makes Name optional.
	// If Name is actually required, you can use input.Name directly.
	if input.Name == nil {
		return nil, fmt.Errorf("project name is required")
	}

	var description pgtype.Text

	if input.Description != nil {
		description = pgtype.Text{
			String: *input.Description,
			Valid:  true,
		}
	}

	project, err := r.App.Services.Project.UpdateProject(
		ctx,
		database.UpdateProjectParams{
			ID:          projectID,
			Name:        *input.Name,
			Description: description,
		},
	)
	if err != nil {
		return nil, err
	}

	result := helpers.ProjectToModel(project)

	// Project.Owner is required in GraphQL.
	user, err := r.App.Services.User.GetUserByID(ctx, project.OwnerID)
	if err != nil {
		return nil, err
	}

	result.Owner = &model.User{
		ID:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	}

	return result, nil
}

// ArchiveProject is the resolver for the archiveProject field.
func (r *mutationResolver) ArchiveProject(ctx context.Context, projectID string) (bool, error) {
	id, err := parseUUID(projectID)
	if err != nil {
		return false, err
	}

	if err := r.App.Services.Project.ArchiveProject(ctx, id); err != nil {
		return false, err
	}

	return true, nil
}

// RestoreProject is the resolver for the restoreProject field.
func (r *mutationResolver) RestoreProject(ctx context.Context, projectID string) (bool, error) {
	id, err := parseUUID(projectID)
	if err != nil {
		return false, err
	}

	if err := r.App.Services.Project.RestoreProject(ctx, id); err != nil {
		return false, err
	}

	return true, nil
}

// DeleteProject is the resolver for the deleteProject field.
func (r *mutationResolver) DeleteProject(ctx context.Context, id string) (bool, error) {
	projectID, err := parseUUID(id)
	if err != nil {
		return false, err
	}

	if err := r.App.Services.Project.DeleteProject(ctx, projectID); err != nil {
		return false, err
	}

	return true, nil
}

// TransferProjectOwnership is the resolver for the transferProjectOwnership field.
func (r *mutationResolver) TransferProjectOwnership(ctx context.Context, input model.TransferProjectOwnershipInput) (*model.Project, error) {
	projectID, err := parseUUID(input.ProjectID)
	if err != nil {
		return nil, err
	}

	ownerID, err := parseUUID(input.OwnerID)
	if err != nil {
		return nil, err
	}

	project, err := r.App.Services.Project.TransferOwnership(
		ctx,
		database.TransferOwnershipParams{
			ID:      projectID,
			OwnerID: ownerID,
		},
	)
	if err != nil {
		return nil, err
	}

	result := helpers.ProjectToModel(project)

	user, err := r.App.Services.User.GetUserByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	result.Owner = &model.User{
		ID:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	}

	return result, nil
}

// AddProjectMember is the resolver for the addProjectMember field.
func (r *mutationResolver) AddProjectMember(ctx context.Context, input model.AddProjectMemberInput) (*model.ProjectMember, error) {
	projectID, err := parseUUID(input.ProjectID)
	if err != nil {
		return nil, err
	}

	userID, err := parseUUID(input.UserID)
	if err != nil {
		return nil, err
	}

	member, err := r.App.Services.Project.AddProjectMember(
		ctx,
		database.AddProjectMemberParams{
			ProjectID: projectID,
			UserID:    userID,
			Role:      database.ProjectRole(input.Role),
		},
	)
	if err != nil {
		return nil, err
	}

	return helpers.ProjectMemberToModel(member), nil
}

// UpdateProjectMemberRole is the resolver for the updateProjectMemberRole field.
func (r *mutationResolver) UpdateProjectMemberRole(ctx context.Context, input model.UpdateProjectMemberRoleInput) (*model.ProjectMember, error) {
	projectID, err := parseUUID(input.ProjectID)
	if err != nil {
		return nil, err
	}

	userID, err := parseUUID(input.UserID)
	if err != nil {
		return nil, err
	}

	member, err := r.App.Services.Project.UpdateProjectMemberRole(
		ctx,
		database.UpdateProjectMemberRoleParams{
			ProjectID: projectID,
			UserID:    userID,
			Role:      database.ProjectRole(input.Role),
		},
	)
	if err != nil {
		return nil, err
	}

	return helpers.ProjectMemberToModel(member), nil
}

// RemoveProjectMember is the resolver for the removeProjectMember field.
func (r *mutationResolver) RemoveProjectMember(ctx context.Context, projectID string, userID string) (bool, error) {
	pID, err := parseUUID(projectID)
	if err != nil {
		return false, err
	}

	uID, err := parseUUID(userID)
	if err != nil {
		return false, err
	}

	err = r.App.Services.Project.RemoveProjectMember(
		ctx,
		database.RemoveProjectMemberParams{
			ProjectID: pID,
			UserID:    uID,
		},
	)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Project is the resolver for the project field.
func (r *queryResolver) Project(ctx context.Context, id string) (*model.Project, error) {
	projectID, err := parseUUID(id)
	if err != nil {
		return nil, err
	}

	project, err := r.App.Services.Project.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	result := helpers.ProjectToModel(project)

	// Project.Owner is required.
	user, err := r.App.Services.User.GetUserByID(ctx, project.OwnerID)
	if err != nil {
		return nil, err
	}

	result.Owner = &model.User{
		ID:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
	}

	return result, nil
}

// Projects is the resolver for the projects field.
func (r *queryResolver) Projects(ctx context.Context) ([]*model.Project, error) {
	userId, ok := auth.UserIDFromContext(ctx)
	if !ok {
		fmt.Printf("Error")
	}
	projects, err := r.App.Services.Project.GetProjectsForUser(ctx, userId)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Project, 0, len(projects))

	for _, project := range projects {
		result = append(result, helpers.ProjectToModel(project))
	}

	return result, nil
}

// ProjectsByOwner is the resolver for the projectsByOwner field.
func (r *queryResolver) ProjectsByOwner(ctx context.Context, ownerID string) ([]*model.Project, error) {
	id, err := parseUUID(ownerID)
	if err != nil {
		return nil, err
	}

	projects, err := r.App.Services.Project.GetProjectsByOwnerID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Project, 0, len(projects))

	for _, project := range projects {
		result = append(result, helpers.ProjectToModel(project))
	}

	return result, nil
}

// ProjectsForUser is the resolver for the projectsForUser field.
func (r *queryResolver) ProjectsForUser(ctx context.Context, userID string) ([]*model.Project, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}

	projects, err := r.App.Services.Project.GetProjectsForUser(ctx, id)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Project, 0, len(projects))

	for _, project := range projects {
		result = append(result, helpers.ProjectToModel(project))
	}

	return result, nil
}

// ArchivedProjectsByOwner is the resolver for the archivedProjectsByOwner field.
func (r *queryResolver) ArchivedProjectsByOwner(ctx context.Context, ownerID string) ([]*model.Project, error) {
	id, err := parseUUID(ownerID)
	if err != nil {
		return nil, err
	}

	projects, err := r.App.Services.Project.GetArchivedProjectsByOwner(ctx, id)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Project, 0, len(projects))

	for _, project := range projects {
		result = append(result, helpers.ProjectToModel(project))
	}

	return result, nil
}

// ArchivedProjectsForUser is the resolver for the archivedProjectsForUser field.
func (r *queryResolver) ArchivedProjectsForUser(ctx context.Context, userID string) ([]*model.Project, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}

	projects, err := r.App.Services.Project.GetArchivedProjectsForUser(ctx, id)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Project, 0, len(projects))

	for _, project := range projects {
		result = append(result, helpers.ProjectToModel(project))
	}

	return result, nil
}

// ProjectMembers is the resolver for the projectMembers field.
func (r *queryResolver) ProjectMembers(ctx context.Context, projectID string) ([]*model.ProjectMember, error) {
	id, err := parseUUID(projectID)
	if err != nil {
		return nil, err
	}

	members, err := r.App.Services.Project.GetProjectMembers(ctx, id)
	if err != nil {
		return nil, err
	}

	result := make([]*model.ProjectMember, 0, len(members))

	for _, member := range members {
		result = append(result, helpers.ProjectMemberToModel(member))
	}

	return result, nil
}
