package services

import (
	"context"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/guards"
	"example/hello/internal/repository"
	"example/hello/internal/validators"
)

type ProjectService struct {
	repo         *repository.ProjectRepository
	userGuard    *guards.UserGuard
	projectGuard *guards.ProjectGuard
}

func NewProjectService(
	repo *repository.ProjectRepository,
	userRepo *repository.UserRepository,
) *ProjectService {
	return &ProjectService{
		repo:         repo,
		userGuard:    guards.NewUserGuard(userRepo),
		projectGuard: guards.NewProjectGuard(repo),
	}
}

func (s *ProjectService) CreateProject(
	ctx context.Context,
	params database.CreateProjectParams,
) (database.Project, error) {
	if err := validators.ValidateProjectName(params.Name); err != nil {
		return database.Project{}, err
	}
	if err := validators.ValidateProjectDescription(params.Description.String); err != nil {
		return database.Project{}, err
	}

	if _, err := s.userGuard.EnsureUserExists(ctx, params.OwnerID); err != nil {
		return database.Project{}, err
	}

	if err := s.projectGuard.EnsureProjectNameAvailableForOwner(ctx, params.OwnerID, params.Name); err != nil {
		return database.Project{}, err
	}

	project, err := s.repo.CreateProject(ctx, params)
	if err != nil {
		return database.Project{}, apperrors.InternalError("failed to create project", err)
	}

	return project, nil
}

func (s *ProjectService) AddProjectMember(
	ctx context.Context,
	params database.AddProjectMemberParams,
) (database.ProjectMember, error) {
	if err := validators.ValidateUUID("project id", params.ProjectID); err != nil {
		return database.ProjectMember{}, err
	}
	if err := validators.ValidateUUID("user id", params.UserID); err != nil {
		return database.ProjectMember{}, err
	}

	project, err := s.projectGuard.EnsureProjectExists(ctx, params.ProjectID)
	if err != nil {
		return database.ProjectMember{}, err
	}
	if project.ArchivedAt.Valid {
		return database.ProjectMember{}, apperrors.ForbiddenError("cannot add members to an archived project")
	}

	if _, err := s.userGuard.EnsureUserExists(ctx, params.UserID); err != nil {
		return database.ProjectMember{}, err
	}

	if _, err := s.projectGuard.EnsureProjectMember(ctx, params.ProjectID, params.UserID); err == nil {
		return database.ProjectMember{}, apperrors.ConflictError("user is already a project member")
	} else if !apperrors.IsCode(err, apperrors.Forbidden) {
		return database.ProjectMember{}, err
	}

	member, err := s.repo.AddProjectMember(ctx, params)
	if err != nil {
		return database.ProjectMember{}, apperrors.InternalError("failed to add project member", err)
	}

	return member, nil
}
