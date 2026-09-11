package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"example/hello/internal/database"
	"example/hello/internal/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

type SessionService struct {
	sessionRepo *repository.SessionRepository
	duration    time.Duration
}

func NewSessionService(
	sessionRepo *repository.SessionRepository,
	duration time.Duration,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		duration:    duration,
	}
}

func (s *SessionService) CreateSession(
	ctx context.Context,
	userID pgtype.UUID,
) (database.Session, error) {

	sessionID, err := randomSessionID()
	if err != nil {
		return database.Session{}, err
	}
	fmt.Printf("Generated session ID: %s\n", sessionID.Bytes)
	expiresAt := time.Now().Add(s.duration)

	return s.sessionRepo.CreateSession(
		ctx,
		sessionID,
		userID,
		expiresAt,
	)
}

func (s *SessionService) GetSession(
	ctx context.Context,
	sessionID pgtype.UUID,
) (database.Session, error) {

	session, err := s.sessionRepo.GetValidSession(ctx, sessionID)
	if err != nil {
		return database.Session{}, err
	}

	// if session.ExpiresAt.(time.Time).Before(time.Now()) {
	// 	_ = s.sessionRepo.DeleteSession(ctx, session.ID)

	// 	return database.Session{}, ErrSessionExpired
	// }

	return session, nil
}

func (s *SessionService) DeleteSession(
	ctx context.Context,
	sessionID pgtype.UUID,
) error {
	return s.sessionRepo.DeleteSession(ctx, sessionID)
}

func (s *SessionService) DeleteUserSessions(
	ctx context.Context,
	userID pgtype.UUID,
) error {
	return s.sessionRepo.DeleteSession(ctx, userID)
}

func randomSessionID() (pgtype.UUID, error) {
	var bytes [16]byte

	if _, err := rand.Read(bytes[:]); err != nil {
		return pgtype.UUID{}, err
	}

	return pgtype.UUID{
		Bytes: bytes,
		Valid: true,
	}, nil
}
