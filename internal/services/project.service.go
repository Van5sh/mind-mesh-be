package services

import (
	"context"
	"example/hello/internal/database"
	"example/hello/internal/repository"
)

type ProjectService struct {
	repo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{
		repo: repo,
	}
}

func (s *ProjectService) CreateProject(ctx context.Context, params database.CreateProjectParams) (database.Project, error) {
	projectName:= params.Name
	description := params.Description
	ownerID := params.OwnerID
	
}
