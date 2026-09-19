package repository

import (
	"context"
	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

type ReportRepository struct {
	q *database.Queries
}

func NewReportRepository(q *database.Queries) *ReportRepository {
	return &ReportRepository{
		q: q,
	}
}

func (r *ReportRepository) CreateReport(ctx context.Context, params database.CreateReportParams) (database.CreateReportRow, error) {
	return r.q.CreateReport(ctx, params)
}

func (r *ReportRepository) DeleteReport(ctx context.Context, id pgtype.UUID) error {
	return r.q.DeleteReport(ctx, id)
}

func (r *ReportRepository) GetAIReports(ctx context.Context, projectId pgtype.UUID) ([]database.GetAIReportsRow, error) {
	return r.q.GetAIReports(ctx, projectId)
}

func (r *ReportRepository) GetReportByID(ctx context.Context, id pgtype.UUID) (database.GetReportByIDRow, error) {
	return r.q.GetReportByID(ctx, id)
}

func (r *ReportRepository) GetReportWithProject(ctx context.Context, id pgtype.UUID) (database.GetReportWithProjectRow, error) {
	return r.q.GetReportWithProject(ctx, id)
}

func (r *ReportRepository) GetReportsByChatID(ctx context.Context, sourceChatId pgtype.UUID) ([]database.GetReportsByChatIDRow, error) {
	return r.q.GetReportsByChatID(ctx, sourceChatId)
}

func (r *ReportRepository) GetReportsByFormat(ctx context.Context, params database.GetReportsByFormatParams) ([]database.GetReportsByFormatRow, error) {
	return r.q.GetReportsByFormat(ctx, params)
}

func (r *ReportRepository) GetReportsByGenerator(ctx context.Context, params database.GetReportsByGeneratorParams) ([]database.GetReportsByGeneratorRow, error) {
	return r.q.GetReportsByGenerator(ctx, params)
}

func (r *ReportRepository) GetReportsByProjectID(ctx context.Context, projectId pgtype.UUID) ([]database.GetReportsByProjectIDRow, error) {
	return r.q.GetReportsByProjectID(ctx, projectId)
}

func (r *ReportRepository) GetReportsByStatus(ctx context.Context, params database.GetReportsByStatusParams) ([]database.GetReportsByStatusRow, error) {
	return r.q.GetReportsByStatus(ctx, params)
}

func (r *ReportRepository) UpdateReport(ctx context.Context, params database.UpdateReportParams) (database.UpdateReportRow, error) {
	return r.q.UpdateReport(ctx, params)
}

func (r *ReportRepository) UpdateReportStatus(ctx context.Context, params database.UpdateReportStatusParams) (database.UpdateReportStatusRow, error) {
	return r.q.UpdateReportStatus(ctx, params)
}

func (r *ReportRepository) GetReportsByProjectIDs(ctx context.Context, projectIDs []pgtype.UUID) ([]database.GetReportsByProjectIDsRow, error) {
	return r.q.GetReportsByProjectIDs(ctx, projectIDs)
}
