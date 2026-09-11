package repository

import (
	"context"
	"time"

	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

type SessionRepository struct {
	queries *database.Queries
}

func NewSessionRepository(queries *database.Queries) *SessionRepository {
	return &SessionRepository{
		queries: queries,
	}
}

// CreateSession creates a new authenticated session for a user.
func (r *SessionRepository) CreateSession(
	ctx context.Context,
	id pgtype.UUID,
	userID pgtype.UUID,
	expiresAt time.Time,
) (database.Session, error) {
	return r.queries.CreateSession(
		ctx,
		database.CreateSessionParams{
			ID:        id,
			UserID:    userID,
			ExpiresAt: expiresAt,
		},
	)
}

// GetSessionByID retrieves a session regardless of whether it has expired.
func (r *SessionRepository) GetSessionByID(
	ctx context.Context,
	id pgtype.UUID,
) (database.Session, error) {
	return r.queries.GetSessionByID(ctx, id)
}

// GetValidSession retrieves a session only if it has not expired.
func (r *SessionRepository) GetValidSession(
	ctx context.Context,
	id pgtype.UUID,
) (database.Session, error) {
	return r.queries.GetValidSession(ctx, id)
}

// GetSessionsByUserID retrieves all sessions belonging to a user.
func (r *SessionRepository) GetSessionsByUserID(
	ctx context.Context,
	userID pgtype.UUID,
) ([]database.Session, error) {
	return r.queries.GetSessionsByUserID(ctx, userID)
}

// DeleteSession deletes a single session.
func (r *SessionRepository) DeleteSession(
	ctx context.Context,
	id pgtype.UUID,
) error {
	return r.queries.DeleteSession(ctx, id)
}

// DeleteSessionsByUserID deletes all sessions belonging to a user.
// Useful for "logout from all devices".
func (r *SessionRepository) DeleteSessionsByUserID(
	ctx context.Context,
	userID pgtype.UUID,
) error {
	return r.queries.DeleteSessionsByUserID(ctx, userID)
}

// DeleteExpiredSessions removes all sessions that have expired.
// This can later be called periodically as a cleanup job.
func (r *SessionRepository) DeleteExpiredSessions(
	ctx context.Context,
) error {
	return r.queries.DeleteExpiredSessions(ctx)
}
