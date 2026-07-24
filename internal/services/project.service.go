package services

import (
	"context"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/guards"
	"example/hello/internal/repository"
	"example/hello/internal/validators"

	"github.com/jackc/pgx/v5/pgtype"
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
	} else if !apperrors.IsCode(err, apperrors.Forbidden) && !apperrors.IsCode(err, apperrors.Conflict) {
		return database.ProjectMember{}, err
	}

	member, err := s.repo.AddProjectMember(ctx, params)
	if err != nil {
		return database.ProjectMember{}, apperrors.InternalError("failed to add project member", err)
	}

	return member, nil
}

func (s *ProjectService) ArchiveProject(ctx context.Context, projectID pgtype.UUID) error {
	checkErr := validators.ValidateUUID("project id", projectID)
	if checkErr != nil {
		return apperrors.Validation("Invalid project ID")
	}
	_, err := s.projectGuard.EnsureProjectExists(ctx, projectID)
	if err != nil {
		return apperrors.Validation("Already exists")
	}
	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return apperrors.InternalError("failed to get project", err)
	}
	if project.ArchivedAt.Valid {
		return apperrors.ForbiddenError("project is already archived")
	}

	err = s.repo.ArchiveProject(ctx, projectID)
	if err != nil {
		return apperrors.InternalError("failed to archive project", err)
	}
	return nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, projectID pgtype.UUID) error {
	checkErr := validators.ValidateUUID("project id", projectID)
	if checkErr != nil {
		return apperrors.Validation("Invalid project ID")
	}
	_, err := s.projectGuard.EnsureProjectExists(ctx, projectID)
	if err != nil {
		return apperrors.Validation("Already exists")
	}
	err = s.repo.DeleteProject(ctx, projectID)
	if err != nil {
		return apperrors.InternalError("failed to delete project", err)
	}
	return nil
}

func (s *ProjectService) GetProjectByID(ctx context.Context, projectID pgtype.UUID) (database.Project, error) {
	checkErr := validators.ValidateUUID("project id", projectID)
	if checkErr != nil {
		return database.Project{}, apperrors.Validation("Invalid project ID")
	}
	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return database.Project{}, apperrors.InternalError("failed to get project", err)
	}
	return project, nil
}

func (s *ProjectService) GetProjectMembers(ctx context.Context, projectID pgtype.UUID) ([]database.ProjectMember, error) {
	checkErr := validators.ValidateUUID("project id", projectID)
	if checkErr != nil {
		return nil, apperrors.Validation("Invalid project ID")
	}
	members, err := s.repo.GetProjectMembers(ctx, projectID)
	if err != nil {
		return nil, apperrors.InternalError("failed to get project members", err)
	}
	return members, nil
}

// ...existing code...

func (s *ProjectService) GetProjectsByOwnerID(ctx context.Context, ownerID pgtype.UUID) ([]database.Project, error) {
    if err := validators.ValidateUUID("owner id", ownerID); err != nil {
        return nil, err
    }
    projects, err := s.repo.GetProjectsByOwnerID(ctx, ownerID)
    if err != nil {
        return nil, apperrors.InternalError("failed to get projects by owner", err)
    }
    return projects, nil
}

func (s *ProjectService) GetProjectsForUser(ctx context.Context, userID pgtype.UUID) ([]database.Project, error) {
    if err := validators.ValidateUUID("user id", userID); err != nil {
        return nil, err
    }
    projects, err := s.repo.GetProjectsForUser(ctx, userID)
    if err != nil {
        return nil, apperrors.InternalError("failed to get projects for user", err)
    }
    return projects, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, params database.UpdateProjectParams) (database.Project, error) {
    if err := validators.ValidateUUID("project id", params.ID); err != nil {
        return database.Project{}, err
    }
    project, err := s.repo.UpdateProject(ctx, params)
    if err != nil {
        return database.Project{}, apperrors.InternalError("failed to update project", err)
    }
    return project, nil
}

func (s *ProjectService) RemoveProjectMember(ctx context.Context, params database.RemoveProjectMemberParams) error {
    if err := validators.ValidateUUID("project id", params.ProjectID); err != nil {
        return err
    }
    if err := validators.ValidateUUID("user id", params.UserID); err != nil {
        return err
    }
    err := s.repo.RemoveProjectMember(ctx, params)
    if err != nil {
        return apperrors.InternalError("failed to remove project member", err)
    }
    return nil
}

func (s *ProjectService) UpdateProjectMemberRole(ctx context.Context, params database.UpdateProjectMemberRoleParams) (database.ProjectMember, error) {
    if err := validators.ValidateUUID("project id", params.ProjectID); err != nil {
        return database.ProjectMember{}, err
    }
    if err := validators.ValidateUUID("user id", params.UserID); err != nil {
        return database.ProjectMember{}, err
    }
    member, err := s.repo.UpdateProjectMemberRole(ctx, params)
    if err != nil {
        return database.ProjectMember{}, apperrors.InternalError("failed to update project member role", err)
    }
    return member, nil
}

func (s *ProjectService) RestoreProject(ctx context.Context, projectID pgtype.UUID) error {
    if err := validators.ValidateUUID("project id", projectID); err != nil {
        return err
    }
    err := s.repo.RestoreProject(ctx, projectID)
    if err != nil {
        return apperrors.InternalError("failed to restore project", err)
    }
    return nil
}

func (s *ProjectService) TransferOwnership(ctx context.Context, params database.TransferOwnershipParams) (database.Project, error) {
    if err := validators.ValidateUUID("project id", params.ID); err != nil {
        return database.Project{}, err
    }
    if err := validators.ValidateUUID("new owner id", params.OwnerID); err != nil {
        return database.Project{}, err
    }
    project, err := s.repo.TransferOwnership(ctx, params)
    if err != nil {
        return database.Project{}, apperrors.InternalError("failed to transfer ownership", err)
    }
    return project, nil
}

func (s *ProjectService) GetArchivedProjectsByOwner(ctx context.Context, ownerID pgtype.UUID) ([]database.Project, error) {
    if err := validators.ValidateUUID("owner id", ownerID); err != nil {
        return nil, err
    }
    projects, err := s.repo.GetArchivedProjectsByOwner(ctx, ownerID)
    if err != nil {
        return nil, apperrors.InternalError("failed to get archived projects by owner", err)
    }
    return projects, nil
}

func (s *ProjectService) GetArchivedProjectsForUser(ctx context.Context, userID pgtype.UUID) ([]database.Project, error) {
    if err := validators.ValidateUUID("user id", userID); err != nil {
        return nil, err
    }
    projects, err := s.repo.GetArchivedProjectsForUser(ctx, userID)
    if err != nil {
        return nil, apperrors.InternalError("failed to get archived projects for user", err)
    }
    return projects, nil
}