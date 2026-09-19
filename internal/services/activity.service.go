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

type ActivityService struct {
	repo  *repository.ActivityRepository
	guard *guards.ActivityGuard
}

func NewActivityService(
	repo *repository.ActivityRepository,
	guard *guards.ActivityGuard,
) *ActivityService {
	return &ActivityService{
		repo:  repo,
		guard: guard,
	}
}

func (s *ActivityService) CreateActivityLog(
	ctx context.Context,
	params database.CreateActivityLogParams,
) (database.ActivityLog, error) {
	activity, err := s.repo.CreateActivityLog(ctx, params)
	if err != nil {
		return database.ActivityLog{}, apperrors.InternalError(
			"failed to create activity log",
			err,
		)
	}
	return activity, nil
}

func (s *ActivityService) GetActivityLogByID(
	ctx context.Context,
	id pgtype.UUID,
) (database.ActivityLog, error) {
	if err := validators.ValidateUUID("activity_id", id); err != nil {
		return database.ActivityLog{}, err
	}
	activity, err := s.guard.EnsureActivityLogExists(ctx, id)
	if err != nil {
		return database.ActivityLog{}, err
	}
	return activity, nil
}

func (s *ActivityService) DeleteActivityLogByID(
	ctx context.Context,
	id pgtype.UUID,
) error {
	if err := validators.ValidateUUID("activity_id", id); err != nil {
		return err
	}
	if _, err := s.guard.EnsureActivityLogExists(ctx, id); err != nil {
		return err
	}
	if err := s.repo.DeleteActivityLogByID(ctx, id); err != nil {
		return apperrors.InternalError(
			"failed to delete activity log",
			err,
		)
	}
	return nil
}

func (s *ActivityService) DeleteActivityLogsByProjectID(
	ctx context.Context,
	projectID pgtype.UUID,
) error {
	if err := validators.ValidateUUID("project_id", projectID); err != nil {
		return err
	}
	if err := s.repo.DeleteActivityLogsByProjectID(ctx, projectID); err != nil {
		return apperrors.InternalError(
			"failed to delete project activity logs",
			err,
		)
	}
	return nil
}

func (s *ActivityService) GetActivityLogsByAction(
	ctx context.Context,
	action string,
) ([]database.ActivityLog, error) {
	if err := validators.ValidateActivityAction(action); err != nil {
		return nil, err
	}
	activities, err := s.repo.GetActivityLogsByAction(ctx, action)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch activity logs by action",
			err,
		)
	}
	return activities, nil
}

func (s *ActivityService) GetActivityLogsByEntity(
	ctx context.Context,
	params database.GetActivityLogsByEntityParams,
) ([]database.ActivityLog, error) {
	activities, err := s.repo.GetActivityLogsByEntity(ctx, params)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch activity logs by entity",
			err,
		)
	}
	return activities, nil
}

func (s *ActivityService) GetActivityLogsByProjectAndAction(
	ctx context.Context,
	params database.GetActivityLogsByProjectAndActionParams,
) ([]database.ActivityLog, error) {
	activities, err := s.repo.GetActivityLogsByProjectAndAction(ctx, params)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch activity logs by project and action",
			err,
		)
	}
	return activities, nil
}

func (s *ActivityService) GetActivityLogsByProjectAndUser(
	ctx context.Context,
	params database.GetActivityLogsByProjectAndUserParams,
) ([]database.ActivityLog, error) {
	activities, err := s.repo.GetActivityLogsByProjectAndUser(ctx, params)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch activity logs by project and user",
			err,
		)
	}
	return activities, nil
}

func (s *ActivityService) GetActivityLogsByProjectID(
	ctx context.Context,
	projectID pgtype.UUID,
) ([]database.ActivityLog, error) {
	if err := validators.ValidateUUID("project_id", projectID); err != nil {
		return nil, err
	}
	activities, err := s.repo.GetActivityLogsByProjectID(ctx, projectID)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch project activity logs",
			err,
		)
	}
	return activities, nil
}

func (s *ActivityService) GetActivityLogsByProjectUserAndAction(
	ctx context.Context,
	params database.GetActivityLogsByProjectUserAndActionParams,
) ([]database.ActivityLog, error) {
	activities, err := s.repo.GetActivityLogsByProjectUserAndAction(ctx, params)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch activity logs by project, user and action",
			err,
		)
	}
	return activities, nil
}

func (s *ActivityService) GetActivityLogsByUserAndAction(
	ctx context.Context,
	params database.GetActivityLogsByUserAndActionParams,
) ([]database.ActivityLog, error) {
	activities, err := s.repo.GetActivityLogsByUserAndAction(ctx, params)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch activity logs by user and action",
			err,
		)
	}
	return activities, nil
}

func (s *ActivityService) GetActivityLogsByUserID(
	ctx context.Context,
	userID pgtype.UUID,
) ([]database.ActivityLog, error) {
	if err := validators.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	activities, err := s.repo.GetActivityLogsByUserID(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch user activity logs",
			err,
		)
	}
	return activities, nil
}

func (s *ActivityService) GetActivityLogsPaginated(
	ctx context.Context,
	params database.GetActivityLogsPaginatedParams,
) ([]database.ActivityLog, error) {
	activities, err := s.repo.GetActivityLogsPaginated(ctx, params)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch paginated activity logs",
			err,
		)
	}
	return activities, nil
}

func (s *ActivityService) CountActivityLogsByProjectID(
	ctx context.Context,
	projectID pgtype.UUID,
) (int64, error) {
	if err := validators.ValidateUUID("project_id", projectID); err != nil {
		return 0, err
	}
	count, err := s.repo.CountActivityLogsByProjectID(ctx, projectID)
	if err != nil {
		return 0, apperrors.InternalError(
			"failed to count project activity logs",
			err,
		)
	}
	return count, nil
}

func (s *ActivityService) CountActivityLogsByUserID(
	ctx context.Context,
	userID pgtype.UUID,
) (int64, error) {
	if err := validators.ValidateUUID("user_id", userID); err != nil {
		return 0, err
	}
	count, err := s.repo.CountActivityLogsByUserID(ctx, userID)
	if err != nil {
		return 0, apperrors.InternalError(
			"failed to count user activity logs",
			err,
		)
	}
	return count, nil
}

func (s *ActivityService) GetProjectStats(
	ctx context.Context,
	projectID pgtype.UUID,
) (database.GetProjectStatsRow, error) {
	if err := validators.ValidateUUID("project_id", projectID); err != nil {
		return database.GetProjectStatsRow{}, err
	}
	stats, err := s.repo.GetProjectStats(ctx, projectID)
	if err != nil {
		return database.GetProjectStatsRow{}, apperrors.InternalError(
			"failed to fetch project statistics",
			err,
		)
	}
	return stats, nil
}

func (s *ActivityService) GetRecentActivityLogs(
	ctx context.Context,
	limit int32,
) ([]database.ActivityLog, error) {
	if limit <= 0 {
		return nil, apperrors.Validation(
			"activity log limit must be greater than zero",
		)
	}
	activities, err := s.repo.GetRecentActivityLogs(ctx, limit)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch recent activity logs",
			err,
		)
	}
	return activities, nil
}

// GetActivityLogsByProjectIDs fetches the activity logs of many projects in
// one query, grouped by project ID. It exists for the Project.activityLogs
// dataloader.
func (s *ActivityService) GetActivityLogsByProjectIDs(
	ctx context.Context,
	projectIDs []pgtype.UUID,
) (map[pgtype.UUID][]database.ActivityLog, error) {
	return fetchGrouped(ctx, projectIDs, "activity logs by projects",
		s.repo.GetActivityLogsByProjectIDs,
		func(a database.ActivityLog) pgtype.UUID { return a.ProjectID })
}
