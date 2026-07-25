package services

import "example/hello/internal/repository"

type ReportService struct {
	repo *repository.ReportRepository
}

func NewReportService(repo *repository.ReportRepository) *ReportService {
	return &ReportService{
		repo: repo,
	}
}
