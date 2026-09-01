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

type FlowchartService struct {
	repo  *repository.FlowchartRepository
	guard *guards.FlowchartGuard
}

func NewFlowchartService(
	repo *repository.FlowchartRepository,
	guard *guards.FlowchartGuard,
) *FlowchartService {
	return &FlowchartService{		repo:  repo,
		guard: guard,
	}
}

func (s *FlowchartService) GetFlowchartByID(
	ctx context.Context,
	id pgtype.UUID,
) (database.Flowchart, error) {

	if err := validators.ValidateUUID("flowchart_id", id); err != nil {
		return database.Flowchart{}, err
	}

	flowchart, err := s.guard.EnsureFlowchartExists(ctx, id)
	if err != nil {
		return database.Flowchart{}, err
	}

	return flowchart, nil
}

func (s *FlowchartService) CreateFlowchart(
	ctx context.Context,
	params database.CreateFlowchartParams,
) (database.Flowchart, error) {

	if err := validators.ValidateUUID("project_id", params.ProjectID); err != nil {
		return database.Flowchart{}, err
	}

	if err := validators.ValidateFlowchartName(params.Name); err != nil {
		return database.Flowchart{}, err
	}

	if err := validators.ValidateFlowchartData(params.Data); err != nil {
		return database.Flowchart{}, err
	}

	flowchart, err := s.repo.CreateFlowchart(ctx, params)
	if err != nil {
		return database.Flowchart{}, apperrors.InternalError(
			"failed to create flowchart",
			err,
		)
	}

	return flowchart, nil
}

func (s *FlowchartService) UpdateFlowchart(
	ctx context.Context,
	params database.UpdateFlowchartParams,
) (database.Flowchart, error) {

	if err := validators.ValidateUUID("flowchart_id", params.ID); err != nil {
		return database.Flowchart{}, err
	}

	if err := validators.ValidateFlowchartName(params.Name); err != nil {
		return database.Flowchart{}, err
	}

	if err := validators.ValidateFlowchartData(params.Data); err != nil {
		return database.Flowchart{}, err
	}

	if _, err := s.guard.EnsureFlowchartExists(ctx, params.ID); err != nil {
		return database.Flowchart{}, err
	}

	flowchart, err := s.repo.UpdateFlowchart(ctx, params)
	if err != nil {
		return database.Flowchart{}, apperrors.InternalError(
			"failed to update flowchart",
			err,
		)
	}

	return flowchart, nil
}

func (s *FlowchartService) DeleteFlowchart(
	ctx context.Context,
	id pgtype.UUID,
) error {

	if err := validators.ValidateUUID("flowchart_id", id); err != nil {
		return err
	}

	if _, err := s.guard.EnsureFlowchartExists(ctx, id); err != nil {
		return err
	}

	if err := s.repo.DeleteFlowchart(ctx, id); err != nil {
		return apperrors.InternalError(
			"failed to delete flowchart",
			err,
		)
	}

	return nil
}