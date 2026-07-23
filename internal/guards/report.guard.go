package guards

import (
	"context"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

type ReportGuard struct {
	repo *repository.ReportRepository
}

func NewReportGuard(repo *repository.ReportRepository) *ReportGuard {
	return &ReportGuard{repo: repo}
}

func (g *ReportGuard) EnsureReportExists(ctx context.Context, reportID pgtype.UUID) (database.GetReportByIDRow, error) {
	report, err := g.repo.GetReportByID(ctx, reportID)
	if err != nil {
		if isNoRows(err) {
			return database.GetReportByIDRow{}, apperrors.NotFoundError("report not found")
		}
		return database.GetReportByIDRow{}, apperrors.InternalError("failed to fetch report", err)
	}
	return report, nil
}

func (g *ReportGuard) EnsureReportBelongsToProject(ctx context.Context, reportID, projectID pgtype.UUID) (database.GetReportByIDRow, error) {
	report, err := g.EnsureReportExists(ctx, reportID)
	if err != nil {
		return database.GetReportByIDRow{}, err
	}
	if !sameUUID(report.ProjectID, projectID) {
		return database.GetReportByIDRow{}, apperrors.ForbiddenError("report does not belong to project")
	}
	return report, nil
}
