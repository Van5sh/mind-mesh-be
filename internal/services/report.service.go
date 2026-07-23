package services

import (
	"context"
	"example/hello/internal/repository"
)

type ReportService struct {
	repo *repository.ReportRepository
}

func NewReportService(repo *repository.ReportRepository) *ReportService {
	return &ReportService{
		repo: repo,
	}
}

func (s *ReportService) CreateReport(ctx context.Context)