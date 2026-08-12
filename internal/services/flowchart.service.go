package services

import (
	"context"
	"example/hello/internal/database"
	"example/hello/internal/repository"
	"example/hello/internal/validators"

	"github.com/jackc/pgx/v5/pgtype"
)

type FlowchartService struct {
	repo *repository.FlowchartRepository
}

func NewFlowchartService(repo *repository.FlowchartRepository) *FlowchartService {
	return &FlowchartService{repo: repo}
}

func (s *FlowchartService) GetFlowchartByID(ctx context.Context, id pgtype.UUID) (database.Flowchart, error) {
	err := validators.ValidateUUID("flowchart_id", id)
	if err != nil {
		return database.Flowchart{}, err
	}
	flowchart, err := s.repo.GetFlowchartByID(ctx, id)
	if err != nil {
		return database.Flowchart{}, err
	}
	return flowchart, nil
}

func (s *FlowchartService) CreateFlowchart(
	ctx context.Context,
	params database.CreateFlowchartParams,
) (database.Flowchart, error) {

	err := validators.ValidateUUID("project_id", params.ProjectID)
	if err != nil {
		return database.Flowchart{}, err
	}

	err = validators.ValidateFlowchartName(params.Name)
	if err != nil {
		return database.Flowchart{}, err
	}

	err = validators.ValidateFlowchartData(params.Data)
	if err != nil {
		return database.Flowchart{}, err
	}

	flowchart, err := s.repo.CreateFlowchart(ctx, params)
	if err != nil {
		return database.Flowchart{}, err
	}

	return flowchart, nil
}

func (s *FlowchartService) UpdateFlowchart(ctx context.Context, params database.UpdateFlowchartParams) (database.Flowchart, error) {
	err := validators.ValidateUUID("flowchart_id", params.ID)
	if err != nil {
		return database.Flowchart{}, err
	}
	err = validators.ValidateFlowchartName(params.Name)
	if err != nil {
		return database.Flowchart{}, err
	}
	flowchart, err := s.repo.UpdateFlowchart(ctx, params)
	if err != nil {
		return database.Flowchart{}, err
	}
	return flowchart, nil
}

func (s *FlowchartService) DeleteFlowchart(ctx context.Context, id pgtype.UUID) error {
	err := validators.ValidateUUID("flowchart_id", id)
	if err != nil {
		return err
	}
	err = s.repo.DeleteFlowchart(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
