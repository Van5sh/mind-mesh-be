package helpers

import (
	"example/hello/graph/model"
	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

// ReportRow is satisfied by every sqlc row type that joins reports with
// report_properties (CreateReportRow, GetReportByIDRow, UpdateReportRow, ...).
// It lets a single mapper build the GraphQL model regardless of which query
// produced the row.
type ReportRow struct {
	ID            pgtype.UUID
	ProjectID     pgtype.UUID
	Title         string
	Content       string
	Format        database.ReportFormat
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
	GeneratedBy   pgtype.UUID
	GeneratedByAi pgtype.Bool
	Status        database.ReportStatus
	SourceChatID  pgtype.UUID
}

func ReportToModel(row ReportRow) *model.Report {
	generatedBy := &model.User{}

	if row.GeneratedBy.Valid {
		generatedBy = &model.User{ID: row.GeneratedBy.String()}
	}

	var sourceChat *model.Chat

	if row.SourceChatID.Valid {
		sourceChat = &model.Chat{ID: row.SourceChatID.String()}
	}

	report := &model.Report{
		ID:      row.ID.String(),
		Title:   row.Title,
		Content: row.Content,
		Format:  model.ReportFormat(row.Format),
		Project: &model.Project{
			ID: row.ProjectID.String(),
		},
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}

	report.Properties = &model.ReportProperties{
		Report:        &model.Report{ID: report.ID},
		GeneratedBy:   generatedBy,
		GeneratedByAi: row.GeneratedByAi.Bool,
		Status:        model.ReportStatus(row.Status),
		SourceChat:    sourceChat,
	}

	return report
}
