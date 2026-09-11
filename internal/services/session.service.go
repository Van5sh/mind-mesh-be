package services

import (
	"context"
	"time"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type SessionService struct {
	repo *repository.SessionRepository
}

func NewSessionService(
	repo *repository.SessionRepository,
) *SessionService {
	return &SessionService{
		repo: repo,
	}
}

// CreateSession creates a new server-side session for a user.
func (s *SessionService) CreateSession(
	ctx context.Context,
	userID pgtype.UUID,
) (database.Session, error) {

	sessionID := uuid.New()

	id := pgtype.UUID{
		Bytes: sessionID,
		Valid: true,
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	session, err := s.repo.CreateSession(
		ctx,
		id,
		userID,
		expiresAt,
	)
	if err != nil {
		return database.Session{}, apperrors.InternalError(
			"failed to create session",
			err,
		)
	}

	return session, nil
}

// GetSession returns a session if it exists and has not expired.
func (s *SessionService) GetSession(
	ctx context.Context,
	sessionID pgtype.UUID,
) (database.Session, error) {

	session, err := s.repo.GetValidSession(ctx, sessionID)
	if err != nil {
		return database.Session{}, apperrors.UnauthorizedError(
			"invalid or expired session",
		)
	}

	return session, nil
}

// DeleteSession logs the current session out.
func (s *SessionService) DeleteSession(
	ctx context.Context,
	sessionID pgtype.UUID,
) error {

	if err := s.repo.DeleteSession(ctx, sessionID); err != nil {
		return apperrors.InternalError(
			"failed to delete session",
			err,
		)
	}

	return nil
}

// DeleteUserSessions logs the user out from all devices.
func (s *SessionService) DeleteUserSessions(
	ctx context.Context,
	userID pgtype.UUID,
) error {

	if err := s.repo.DeleteSessionsByUserID(ctx, userID); err != nil {
		return apperrors.InternalError(
			"failed to delete user sessions",
			err,
		)
	}

	return nil
}

// DeleteExpiredSessions removes expired sessions from the database.
func (s *SessionService) DeleteExpiredSessions(
	ctx context.Context,
) error {

	if err := s.repo.DeleteExpiredSessions(ctx); err != nil {
		return apperrors.InternalError(
			"failed to delete expired sessions",
			err,
		)
	}

	return nil
}
