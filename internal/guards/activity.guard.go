package guards

import (
	"context"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

type ActivityGuard struct {
	repo *repository.ActivityRepository
}

func NewActivityGuard(repo *repository.ActivityRepository) *ActivityGuard {
	return &ActivityGuard{repo: repo}
}

func (g *ActivityGuard) EnsureActivityLogExists(ctx context.Context, activityID pgtype.UUID) (database.ActivityLog, error) {
	activity, err := g.repo.GetActivityLogByID(ctx, activityID)
	if err != nil {
		if isNoRows(err) {
			return database.ActivityLog{}, apperrors.NotFoundError("activity log not found")
		}
		return database.ActivityLog{}, apperrors.InternalError("failed to fetch activity log", err)
	}
	return activity, nil
}
