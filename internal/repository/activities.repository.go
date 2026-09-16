package repository

import (
	"context"
	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

type ActivityRepository struct {
	q *database.Queries
}

func NewActivityRepository(q *database.Queries) *ActivityRepository {
	return &ActivityRepository{
		q: q,
	}
}

func (r *ActivityRepository) CountActivityLogsByProjectID(ctx context.Context, projectID pgtype.UUID) (int64, error) {
	return r.q.CountActivityLogsByProjectID(ctx, projectID)
}

func (r *ActivityRepository) CountActivityLogsByUserID(ctx context.Context, userID pgtype.UUID) (int64, error) {
	return r.q.CountActivityLogsByUserID(ctx, userID)
}

func (r *ActivityRepository) CreateActivityLog(ctx context.Context, params database.CreateActivityLogParams) (database.ActivityLog, error) {
	return r.q.CreateActivityLog(ctx, params)
}

func (r *ActivityRepository) DeleteActivityLogByID(ctx context.Context, id pgtype.UUID) error {
	return r.q.DeleteActivityLogByID(ctx, id)
}

func (r *ActivityRepository) DeleteActivityLogsByProjectID(ctx context.Context, projectID pgtype.UUID) error {
	return r.q.DeleteActivityLogsByProjectID(ctx, projectID)
}

func (r *ActivityRepository) GetActivityLogByID(ctx context.Context, id pgtype.UUID) (database.ActivityLog, error) {
	return r.q.GetActivityLogByID(ctx, id)
}

func (r *ActivityRepository) GetActivityLogsByAction(ctx context.Context, action string) ([]database.ActivityLog, error) {
	return r.q.GetActivityLogsByAction(ctx, action)
}

func (r *ActivityRepository) GetActivityLogsByEntity(ctx context.Context, params database.GetActivityLogsByEntityParams) ([]database.ActivityLog, error) {
	return r.q.GetActivityLogsByEntity(ctx, params)
}

func (r *ActivityRepository) GetActivityLogsByProjectAndAction(ctx context.Context, params database.GetActivityLogsByProjectAndActionParams) ([]database.ActivityLog, error) {
	return r.q.GetActivityLogsByProjectAndAction(ctx, params)
}

func (r *ActivityRepository) GetActivityLogsByProjectAndUser(ctx context.Context, params database.GetActivityLogsByProjectAndUserParams) ([]database.ActivityLog, error) {
	return r.q.GetActivityLogsByProjectAndUser(ctx, params)
}

func (r *ActivityRepository) GetActivityLogsByProjectID(ctx context.Context, projectID pgtype.UUID) ([]database.ActivityLog, error) {
	return r.q.GetActivityLogsByProjectID(ctx, projectID)
}

func (r *ActivityRepository) GetActivityLogsByProjectUserAndAction(ctx context.Context, params database.GetActivityLogsByProjectUserAndActionParams) ([]database.ActivityLog, error) {
	return r.q.GetActivityLogsByProjectUserAndAction(ctx, params)
}

func (r *ActivityRepository) GetActivityLogsByUserAndAction(ctx context.Context, params database.GetActivityLogsByUserAndActionParams) ([]database.ActivityLog, error) {
	return r.q.GetActivityLogsByUserAndAction(ctx, params)
}

func (r *ActivityRepository) GetActivityLogsByUserID(ctx context.Context, userID pgtype.UUID) ([]database.ActivityLog, error) {
	return r.q.GetActivityLogsByUserID(ctx, userID)
}

func (r *ActivityRepository) GetActivityLogsPaginated(ctx context.Context, params database.GetActivityLogsPaginatedParams) ([]database.ActivityLog, error) {
	return r.q.GetActivityLogsPaginated(ctx, params)
}

func (r *ActivityRepository) GetProjectStats(ctx context.Context, projectID pgtype.UUID) (database.GetProjectStatsRow, error) {
	return r.q.GetProjectStats(ctx, projectID)
}

func (r *ActivityRepository) GetRecentActivityLogs(ctx context.Context, limit int32) ([]database.ActivityLog, error) {
	return r.q.GetRecentActivityLogs(ctx, limit)
}
