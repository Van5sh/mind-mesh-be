package repository

import (
	"context"
	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

type FlowchartRepository struct {
	q *database.Queries
}

func NewFlowchartRepository(q *database.Queries) *FlowchartRepository {
	return &FlowchartRepository{
		q: q,
	}
}

func (r *FlowchartRepository) CreateFlowchart(ctx context.Context, params database.CreateFlowchartParams) (database.Flowchart, error) {
	return r.q.CreateFlowchart(ctx, params)
}

func (r *FlowchartRepository) DeleteFlowchart(ctx context.Context, id pgtype.UUID) error {
	return r.q.DeleteFlowchart(ctx, id)
}

func (r *FlowchartRepository) GetAIFlowCharts(ctx context.Context) ([]database.Flowchart, error) {
	return r.q.GetAIFlowCharts(ctx)
}

func (r *FlowchartRepository) GetFlowchartByID(ctx context.Context, id pgtype.UUID) (database.Flowchart, error) {
	return r.q.GetFlowchartByID(ctx, id)
}

func (r *FlowchartRepository) GetFlowchartWithProject(ctx context.Context, id pgtype.UUID) (database.GetFlowchartWithProjectRow, error) {
	return r.q.GetFlowchartWithProject(ctx, id)
}

func (r *FlowchartRepository) GetFlowchartsByChatID(ctx context.Context, sourceChatID pgtype.UUID) ([]database.Flowchart, error) {
	return r.q.GetFlowchartsByChatID(ctx, sourceChatID)
}

func (r *FlowchartRepository) GetFlowchartsByGenerator(ctx context.Context, generatedBy pgtype.UUID) ([]database.Flowchart, error) {
	return r.q.GetFlowchartsByGenerator(ctx, generatedBy)
}

func (r *FlowchartRepository) GetFlowchartsByProjectAndStatus(ctx context.Context, params database.GetFlowchartsByProjectAndStatusParams) ([]database.Flowchart, error) {
	return r.q.GetFlowchartsByProjectAndStatus(ctx, params)
}

func (r *FlowchartRepository) GetFlowchartsByProjectID(ctx context.Context, projectID pgtype.UUID) ([]database.Flowchart, error) {
	return r.q.GetFlowchartsByProjectID(ctx, projectID)
}

func (r *FlowchartRepository) GetFlowchartsByStatus(ctx context.Context, status database.FlowchartStatus) ([]database.Flowchart, error) {
	return r.q.GetFlowchartsByStatus(ctx, status)
}

func (r *FlowchartRepository) GetLatestFlowchart(ctx context.Context, projectID pgtype.UUID) (database.Flowchart, error) {
	return r.q.GetLatestFlowchart(ctx, projectID)
}

func (r *FlowchartRepository) RenameFlowchart(ctx context.Context, params database.RenameFlowchartParams) (database.Flowchart, error) {
	return r.q.RenameFlowchart(ctx, params)
}

func (r *FlowchartRepository) UpdateFlowchart(ctx context.Context, params database.UpdateFlowchartParams) (database.Flowchart, error) {
	return r.q.UpdateFlowchart(ctx, params)
}

func (r *FlowchartRepository) UpdateFlowchartStatus(ctx context.Context, params database.UpdateFlowchartStatusParams) (database.Flowchart, error) {
	return r.q.UpdateFlowchartStatus(ctx, params)
}
