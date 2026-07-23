package guards

import (
	"context"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

type FlowchartGuard struct {
	repo *repository.FlowchartRepository
}

func NewFlowchartGuard(repo *repository.FlowchartRepository) *FlowchartGuard {
	return &FlowchartGuard{repo: repo}
}

func (g *FlowchartGuard) EnsureFlowchartExists(ctx context.Context, flowchartID pgtype.UUID) (database.Flowchart, error) {
	flowchart, err := g.repo.GetFlowchartByID(ctx, flowchartID)
	if err != nil {
		if isNoRows(err) {
			return database.Flowchart{}, apperrors.NotFoundError("flowchart not found")
		}
		return database.Flowchart{}, apperrors.InternalError("failed to fetch flowchart", err)
	}
	return flowchart, nil
}

func (g *FlowchartGuard) EnsureFlowchartBelongsToProject(ctx context.Context, flowchartID, projectID pgtype.UUID) (database.Flowchart, error) {
	flowchart, err := g.EnsureFlowchartExists(ctx, flowchartID)
	if err != nil {
		return database.Flowchart{}, err
	}
	if !sameUUID(flowchart.ProjectID, projectID) {
		return database.Flowchart{}, apperrors.ForbiddenError("flowchart does not belong to project")
	}
	return flowchart, nil
}
