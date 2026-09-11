package auth

import (
	"context"
	"fmt"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/repository"
	"example/hello/internal/services"

	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	userRepo   *repository.UserRepository
	oauthRepo  *repository.OAuthRepository
	sessionSvc *services.SessionService
}

func NewService(
	userRepo *repository.UserRepository,
	oauthRepo *repository.OAuthRepository,
	sessionSvc *services.SessionService,
) *Service {
	return &Service{
		userRepo:   userRepo,
		oauthRepo:  oauthRepo,
		sessionSvc: sessionSvc,
	}
}

type OAuthUser struct {
	Provider       string
	ProviderUserID string
	Email          string
	Username       string
}

func (s *Service) LoginWithOAuth(
	ctx context.Context,
	oauthUser OAuthUser,
) (database.User, database.Session, error) {

	account, err := s.oauthRepo.GetOAuthAccount(
		ctx,
		oauthUser.Provider,
		oauthUser.ProviderUserID,
	)

	if err == nil {
		user, err := s.userRepo.GetUserByID(ctx, account.UserID)
		if err != nil {
			return database.User{}, database.Session{},
				apperrors.InternalError(
					"failed to fetch OAuth user",
					err,
				)
		}

		session, err := s.sessionSvc.CreateSession(ctx, user.ID)
		if err != nil {
			return database.User{}, database.Session{}, err
		}

		return user, session, nil
	}

	user, err := s.userRepo.GetUserByEmail(ctx, oauthUser.Email)

	if err == nil {
		_, err = s.oauthRepo.CreateOAuthAccount(
			ctx,
			user.ID,
			oauthUser.Provider,
			oauthUser.ProviderUserID,
		)

		if err != nil {
			return database.User{}, database.Session{},
				apperrors.InternalError(
					"failed to link OAuth account",
					err,
				)
		}

		session, err := s.sessionSvc.CreateSession(ctx, user.ID)
		if err != nil {
			return database.User{}, database.Session{}, err
		}

		return user, session, nil
	}

	username := oauthUser.Username
	username, err = s.generateUniqueUsername(ctx, username)
	if err != nil {
		return database.User{}, database.Session{}, err
	}

	userID := newUUID()

	user, err = s.userRepo.CreateUser(
		ctx,
		database.CreateUserParams{
			ID:       userID,
			Username: username,
			Email:    oauthUser.Email,
		},
	)
	if err != nil {
		return database.User{}, database.Session{},
			apperrors.InternalError(
				"failed to create OAuth user",
				err,
			)
	}

	_, err = s.oauthRepo.CreateOAuthAccount(
		ctx,
		user.ID,
		oauthUser.Provider,
		oauthUser.ProviderUserID,
	)
	if err != nil {
		return database.User{}, database.Session{},
			apperrors.InternalError(
				"failed to create OAuth account",
				err,
			)
	}
	_, err = s.userRepo.CreateUserProfile(
		ctx,
		database.CreateUserProfileParams{
			UserID:    user.ID,
			FirstName: "",
			LastName:  "",
			Bio: pgtype.Text{
				Valid: false,
			},
			AvatarUrl: pgtype.Text{
				Valid: false,
			},
		},
	)
	if err != nil {
		return database.User{}, database.Session{},
			apperrors.InternalError(
				"failed to create OAuth user profile",
				err,
			)
	}
	session, err := s.sessionSvc.CreateSession(ctx, user.ID)
	if err != nil {
		return database.User{}, database.Session{}, err
	}

	return user, session, nil
}

func (s *Service) Logout(
	ctx context.Context,
	sessionID pgtype.UUID,
) error {
	if !sessionID.Valid {
		return nil
	}

	return s.sessionSvc.DeleteSession(ctx, sessionID)
}

func (s *Service) GetUserFromSession(
	ctx context.Context,
	sessionID pgtype.UUID,
) (database.User, error) {

	session, err := s.sessionSvc.GetSession(ctx, sessionID)
	if err != nil {
		return database.User{}, err
	}

	user, err := s.userRepo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return database.User{}, apperrors.UnauthorizedError(
			"invalid session",
		)
	}

	return user, nil
}

func (s *Service) LogoutAll(
	ctx context.Context,
	userID pgtype.UUID,
) error {
	return s.sessionSvc.DeleteUserSessions(ctx, userID)
}

func (s *Service) generateUniqueUsername(
	ctx context.Context,
	username string,
) (string, error) {

	if username == "" {
		username = "user"
	}

	candidate := username

	for i := 1; ; i++ {
		exists, err := s.userRepo.CheckUsernameExists(
			ctx,
			candidate,
		)
		if err != nil {
			return "", apperrors.InternalError(
				"failed to check username",
				err,
			)
		}

		if !exists {
			return candidate, nil
		}

		candidate = fmt.Sprintf("%s%d", username, i+1)
	}
}

func newUUID() pgtype.UUID {
	return pgtype.UUID{
		Bytes: [16]byte{},
		Valid: true,
	}
}
