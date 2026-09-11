package graph

import (
	"context"
	"fmt"

	"example/hello/graph/model"
	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

// ============================================================
// Mutations
// ============================================================

// CreateReport is the resolver for the createReport field.
func (r *mutationResolver) CreateReport(
	ctx context.Context,
	input model.CreateReportInput,
) (*model.Report, error) {

	projectID, err := parseUUID(input.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	generatedByAI := pgtype.Bool{}

	if input.GeneratedByAi != nil {
		generatedByAI = pgtype.Bool{
			Bool:  *input.GeneratedByAi,
			Valid: true,
		}
	}

	report, err := r.App.Services.Report.CreateReport(
		ctx,
		database.CreateReportParams{
			ProjectID:     projectID,
			Title:         input.Title,
			GeneratedByAi: generatedByAI,
			Content:       input.Content,
			Format:        database.ReportFormat(input.Format),
		},
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to create report: %w",
			err,
		)
	}

	return &model.Report{
		ID:         report.ID.String(),
		Properties: &model.ReportProperties{},
		Title:      report.Title,
		Format:     model.ReportFormat(report.Format),
	}, nil
}

// UpdateReport is the resolver for the updateReport field.
func (r *mutationResolver) UpdateReport(
	ctx context.Context,
	id string,
	input model.UpdateReportInput,
) (*model.Report, error) {

	reportID, err := parseUUID(id)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid report ID: %w",
			err,
		)
	}

	report, err := r.App.Services.Report.UpdateReport(
		ctx,
		database.UpdateReportParams{
			ID:      reportID,
			Title:   *input.Title,
			Content: *input.Content,
		},
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to update report: %w",
			err,
		)
	}

	return &model.Report{
		ID:         report.ID.String(),
		Properties: &model.ReportProperties{},
		Title:      report.Title,
		Format:     model.ReportFormat(report.Format),
	}, nil
}

// DeleteReport is the resolver for the deleteReport field.
func (r *mutationResolver) DeleteReport(
	ctx context.Context,
	id string,
) (bool, error) {

	reportID, err := parseUUID(id)
	if err != nil {
		return false, fmt.Errorf(
			"invalid report ID: %w",
			err,
		)
	}

	if err := r.App.Services.Report.DeleteReport(
		ctx,
		reportID,
	); err != nil {
		return false, fmt.Errorf(
			"failed to delete report: %w",
			err,
		)
	}

	return true, nil
}

// ============================================================
// Queries
// ============================================================

// Report is the resolver for the report field.
func (r *queryResolver) Report(
	ctx context.Context,
	id string,
) (*model.Report, error) {

	reportID, err := parseUUID(id)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid report ID: %w",
			err,
		)
	}

	report, err := r.App.Services.Report.GetReportByID(
		ctx,
		reportID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to fetch report: %w",
			err,
		)
	}

	return &model.Report{
		ID:         report.ID.String(),
		Properties: &model.ReportProperties{},
		Title:      report.Title,
		Format:     model.ReportFormat(report.Format),
	}, nil
}

// Reports is the resolver for the reports field.
func (r *queryResolver) Reports(
	ctx context.Context,
	projectID string,
) ([]*model.Report, error) {

	id, err := parseUUID(projectID)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid project ID: %w",
			err,
		)
	}

	reports, err := r.App.Services.Report.GetReportsByProjectID(
		ctx,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to fetch project reports: %w",
			err,
		)
	}

	result := make([]*model.Report, 0, len(reports))

	for _, report := range reports {
		result = append(result, &model.Report{
			ID:         report.ID.String(),
			Properties: &model.ReportProperties{},
			Title:      report.Title,
			Format:     model.ReportFormat(report.Format),
		})
	}

	return result, nil
}

// ReportsByChat is the resolver for the reportsByChat field.
func (r *queryResolver) ReportsByChat(
	ctx context.Context,
	sourceChatID string,
) ([]*model.Report, error) {

	chatID, err := parseUUID(sourceChatID)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid source chat ID: %w",
			err,
		)
	}

	reports, err := r.App.Services.Report.GetReportsByChatID(
		ctx,
		chatID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to fetch reports by chat: %w",
			err,
		)
	}

	result := make([]*model.Report, 0, len(reports))

	for _, report := range reports {
		result = append(result, &model.Report{
			ID:         report.ID.String(),
			Properties: &model.ReportProperties{},
			Title:      report.Title,
			Format:     model.ReportFormat(report.Format),
		})
	}

	return result, nil
}

// ReportsByGenerator is the resolver for the reportsByGenerator field.
func (r *queryResolver) ReportsByGenerator(
	ctx context.Context,
	generatedByID string,
) ([]*model.Report, error) {

	generatorID, err := parseUUID(generatedByID)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid generated-by ID: %w",
			err,
		)
	}

	reports, err := r.App.Services.Report.GetReportsByGenerator(
		ctx,
		generatorID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to fetch reports by generator: %w",
			err,
		)
	}

	result := make([]*model.Report, 0, len(reports))

	for _, report := range reports {
		result = append(result, &model.Report{
			ID:         report.ID.String(),
			Properties: &model.ReportProperties{},
			Title:      report.Title,
			Format:     model.ReportFormat(report.Format),
		})
	}

	return result, nil
}

// ReportsByFormat is the resolver for the reportsByFormat field.
func (r *queryResolver) ReportsByFormat(
	ctx context.Context,
	format model.ReportFormat,
) ([]*model.Report, error) {

	reports, err := r.App.Services.Report.GetReportsByFormat(
		ctx,
		database.GetReportsByFormatParams{
			Format: database.ReportFormat(format),
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to fetch reports by format: %w",
			err,
		)
	}

	result := make([]*model.Report, 0, len(reports))

	for _, report := range reports {
		result = append(result, &model.Report{
			ID:         report.ID.String(),
			Properties: &model.ReportProperties{},
			Title:      report.Title,
			Format:     model.ReportFormat(report.Format),
		})
	}

	return result, nil
}

// ReportsByStatus is the resolver for the reportsByStatus field.
func (r *queryResolver) ReportsByStatus(
	ctx context.Context,
	status model.ReportStatus,
) ([]*model.Report, error) {

	reports, err := r.App.Services.Report.GetReportsByStatus(
		ctx,
		database.GetReportsByStatusParams{
			Status: database.ReportStatus(status),
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to fetch reports by status: %w",
			err,
		)
	}

	result := make([]*model.Report, 0, len(reports))

	for _, report := range reports {
		result = append(result, &model.Report{
			ID:         report.ID.String(),
			Properties: &model.ReportProperties{},
			Title:      report.Title,
			Format:     model.ReportFormat(report.Format),
		})
	}

	return result, nil
}

func (r *queryResolver) AiReports(
	ctx context.Context,
	projectID string,
) ([]*model.Report, error) {

	id, err := parseUUID(projectID)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid project ID: %w",
			err,
		)
	}

	reports, err := r.App.Services.Report.GetAIReports(
		ctx,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to fetch AI reports: %w",
			err,
		)
	}

	result := make([]*model.Report, 0, len(reports))

	for _, report := range reports {
		result = append(result, &model.Report{
			ID:         report.ID.String(),
			Properties: &model.ReportProperties{},
			Title:      report.Title,
			Format:     model.ReportFormat(report.Format),
		})
	}

	return result, nil
}
