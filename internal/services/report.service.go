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

type ReportService struct {
	repo  *repository.ReportRepository
	guard *guards.ReportGuard
}

func NewReportService(
	repo *repository.ReportRepository,
	guard *guards.ReportGuard,
) *ReportService {
	return &ReportService{
		repo:  repo,
		guard: guard,
	}
}

func (s *ReportService) CreateReport(
	ctx context.Context,
	params database.CreateReportParams,
) (database.CreateReportRow, error) {

	if err := validators.ValidateUUID("project_id", params.ProjectID); err != nil {
		return database.CreateReportRow{}, err
	}

	if err := validators.ValidateReportTitle(params.Title); err != nil {
		return database.CreateReportRow{}, err
	}

	if err := validators.ValidateReportContent(params.Content); err != nil {
		return database.CreateReportRow{}, err
	}

	report, err := s.repo.CreateReport(ctx, params)
	if err != nil {
		return database.CreateReportRow{}, apperrors.InternalError(
			"failed to create report",
			err,
		)
	}

	return report, nil
}

func (s *ReportService) GetReportByID(
	ctx context.Context,
	id pgtype.UUID,
) (database.GetReportByIDRow, error) {

	if err := validators.ValidateUUID("report_id", id); err != nil {
		return database.GetReportByIDRow{}, err
	}

	report, err := s.guard.EnsureReportExists(ctx, id)
	if err != nil {
		return database.GetReportByIDRow{}, err
	}

	return report, nil
}

func (s *ReportService) GetReportWithProject(
	ctx context.Context,
	id pgtype.UUID,
) (database.GetReportWithProjectRow, error) {

	if err := validators.ValidateUUID("report_id", id); err != nil {
		return database.GetReportWithProjectRow{}, err
	}

	report, err := s.repo.GetReportWithProject(ctx, id)
	if err != nil {
		return database.GetReportWithProjectRow{}, apperrors.InternalError(
			"failed to fetch report with project",
			err,
		)
	}

	return report, nil
}

func (s *ReportService) DeleteReport(
	ctx context.Context,
	id pgtype.UUID,
) error {

	if err := validators.ValidateUUID("report_id", id); err != nil {
		return err
	}

	if _, err := s.guard.EnsureReportExists(ctx, id); err != nil {
		return err
	}

	if err := s.repo.DeleteReport(ctx, id); err != nil {
		return apperrors.InternalError(
			"failed to delete report",
			err,
		)
	}

	return nil
}

func (s *ReportService) GetAIReports(
	ctx context.Context,
	projectID pgtype.UUID,
) ([]database.GetAIReportsRow, error) {

	if err := validators.ValidateUUID("project_id", projectID); err != nil {
		return nil, err
	}

	reports, err := s.repo.GetAIReports(ctx, projectID)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch AI reports",
			err,
		)
	}

	return reports, nil
}

func (s *ReportService) GetReportsByChatID(
	ctx context.Context,
	sourceChatID pgtype.UUID,
) ([]database.GetReportsByChatIDRow, error) {

	if err := validators.ValidateUUID("source_chat_id", sourceChatID); err != nil {
		return nil, err
	}

	reports, err := s.repo.GetReportsByChatID(ctx, sourceChatID)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch reports by chat",
			err,
		)
	}

	return reports, nil
}

func (s *ReportService) GetReportsByFormat(
	ctx context.Context,
	params database.GetReportsByFormatParams,
) ([]database.GetReportsByFormatRow, error) {

	reports, err := s.repo.GetReportsByFormat(ctx, params)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch reports by format",
			err,
		)
	}

	return reports, nil
}

func (s *ReportService) GetReportsByGenerator(
	ctx context.Context,
	generatedBy pgtype.UUID,
) ([]database.GetReportsByGeneratorRow, error) {

	if err := validators.ValidateUUID("generated_by", generatedBy); err != nil {
		return nil, err
	}

	reports, err := s.repo.GetReportsByGenerator(ctx, generatedBy)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch reports by generator",
			err,
		)
	}

	return reports, nil
}

func (s *ReportService) GetReportsByProjectID(
	ctx context.Context,
	projectID pgtype.UUID,
) ([]database.GetReportsByProjectIDRow, error) {

	if err := validators.ValidateUUID("project_id", projectID); err != nil {
		return nil, err
	}

	reports, err := s.repo.GetReportsByProjectID(ctx, projectID)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch reports by project",
			err,
		)
	}

	return reports, nil
}

func (s *ReportService) GetReportsByStatus(
	ctx context.Context,
	params database.GetReportsByStatusParams,
) ([]database.GetReportsByStatusRow, error) {

	reports, err := s.repo.GetReportsByStatus(ctx, params)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch reports by status",
			err,
		)
	}

	return reports, nil
}

func (s *ReportService) UpdateReport(
	ctx context.Context,
	params database.UpdateReportParams,
) (database.UpdateReportRow, error) {

	if err := validators.ValidateUUID("report_id", params.ID); err != nil {
		return database.UpdateReportRow{}, err
	}

	if err := validators.ValidateReportTitle(params.Title); err != nil {
		return database.UpdateReportRow{}, err
	}

	if err := validators.ValidateReportContent(params.Content); err != nil {
		return database.UpdateReportRow{}, err
	}

	if _, err := s.guard.EnsureReportExists(ctx, params.ID); err != nil {
		return database.UpdateReportRow{}, err
	}

	report, err := s.repo.UpdateReport(ctx, params)
	if err != nil {
		return database.UpdateReportRow{}, apperrors.InternalError(
			"failed to update report",
			err,
		)
	}

	return report, nil
}

func (s *ReportService) UpdateReportStatus(
	ctx context.Context,
	params database.UpdateReportStatusParams,
) (database.UpdateReportStatusRow, error) {

	if err := validators.ValidateUUID("report_id", params.ID); err != nil {
		return database.UpdateReportStatusRow{}, err
	}

	if _, err := s.guard.EnsureReportExists(ctx, params.ID); err != nil {
		return database.UpdateReportStatusRow{}, err
	}

	report, err := s.repo.UpdateReportStatus(ctx, params)
	if err != nil {
		return database.UpdateReportStatusRow{}, apperrors.InternalError(
			"failed to update report status",
			err,
		)
	}
	return report, nil
}
