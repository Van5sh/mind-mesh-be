package services

import (
	"context"

	"example/hello/internal/database"
	"example/hello/internal/guards"
	"example/hello/internal/repository"
)

type FileService struct {
	repo   *repository.FileRepository
	guards *guards.FileGuard
}

func NewFileService(repo *repository.FileRepository) *FileService {
	return &FileService{
		repo:   repo,
		guards: &guards.FileGuard{},
	}
}

func (s *FileService) CreateFile(ctx context.Context, params database.CreateFileParams) (database.File, error) {
	return s.repo.CreateFile(ctx, params)
}
