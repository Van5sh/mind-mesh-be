package auth

import (
	"context"
	"fmt"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/repository"
	"example/hello/internal/services"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	userRepo   *repository.UserRepository
	oauthRepo  *repository.OAuthRepository
	sessionSvc *services.SessionService

	google *GoogleProvider
	github *GitHubProvider
}

func NewService(
	userRepo *repository.UserRepository,
	oauthRepo *repository.OAuthRepository,
	sessionSvc *services.SessionService,
	google *GoogleProvider,
	github *GitHubProvider,
) *Service {
	return &Service{
		userRepo:   userRepo,
		oauthRepo:  oauthRepo,
		sessionSvc: sessionSvc,
		google:     google,
		github:     github,
	}
}

type OAuthUser struct {
	Provider       string
	ProviderUserID string
	Email          string
	Username       string

	FirstName string
	LastName  string
	AvatarURL string
}

// ============================================================
// OAuth Login
// ============================================================

func (s *Service) LoginWithGoogle(
	ctx context.Context,
	code string,
) (database.User, database.Session, error) {

	googleUser, err := s.google.Exchange(ctx, code)
	if err != nil {
		return database.User{}, database.Session{}, err
	}

	oauthUser := OAuthUser{
		Provider:       "GOOGLE",
		ProviderUserID: googleUser.Sub,
		Email:          googleUser.Email,
		Username:       googleUser.GivenName,
		FirstName:      googleUser.GivenName,
		LastName:       googleUser.FamilyName,
		AvatarURL:      googleUser.Picture,
	}

	return s.LoginWithOAuth(ctx, oauthUser)
}

func (s *Service) LoginWithGitHub(
	ctx context.Context,
	code string,
) (database.User, database.Session, error) {

	githubUser, err := s.github.Exchange(ctx, code)
	if err != nil {
		return database.User{}, database.Session{}, err
	}

	oauthUser := OAuthUser{
		Provider:       "GITHUB",
		ProviderUserID: GitHubUserID(githubUser),
		Email:          githubUser.Email,
		Username:       githubUser.Login,
	}

	return s.LoginWithOAuth(ctx, oauthUser)
}

// ============================================================
// Generic OAuth Login
// ============================================================

func (s *Service) LoginWithOAuth(
	ctx context.Context,
	oauthUser OAuthUser,
) (database.User, database.Session, error) {

	// --------------------------------------------------------
	// 1. Check whether OAuth account already exists
	// --------------------------------------------------------

	account, err := s.oauthRepo.GetOAuthAccount(
		ctx,
		oauthUser.Provider,
		oauthUser.ProviderUserID,
	)

	if err == nil {

		user, err := s.userRepo.GetUserByID(
			ctx,
			account.UserID,
		)

		if err != nil {
			return database.User{}, database.Session{},
				apperrors.InternalError(
					"failed to fetch OAuth user",
					err,
				)
		}

		session, err := s.sessionSvc.CreateSession(
			ctx,
			user.ID,
		)

		if err != nil {
			return database.User{}, database.Session{}, err
		}

		return user, session, nil
	}

	// --------------------------------------------------------
	// 2. Check whether user already exists by email
	// --------------------------------------------------------

	user, err := s.userRepo.GetUserByEmail(
		ctx,
		oauthUser.Email,
	)

	if err == nil {

		// Link OAuth account to existing user.
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

		session, err := s.sessionSvc.CreateSession(
			ctx,
			user.ID,
		)

		if err != nil {
			return database.User{}, database.Session{}, err
		}

		return user, session, nil
	}

	// --------------------------------------------------------
	// 3. Create a new user
	// --------------------------------------------------------

	username, err := s.generateUniqueUsername(
		ctx,
		oauthUser.Username,
	)

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

	// --------------------------------------------------------
	// 4. Create OAuth account
	// --------------------------------------------------------

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

	// --------------------------------------------------------
	// 5. Create user profile
	// --------------------------------------------------------

	_, err = s.userRepo.CreateUserProfile(
		ctx,
		database.CreateUserProfileParams{
			UserID:    user.ID,
			FirstName: oauthUser.FirstName,
			LastName:  oauthUser.LastName,
			Bio:       textFromString(""),
			AvatarUrl: textFromString(oauthUser.AvatarURL),
		},
	)

	if err != nil {
		return database.User{}, database.Session{},
			apperrors.InternalError(
				"failed to create OAuth user profile",
				err,
			)
	}

	// --------------------------------------------------------
	// 6. Create session
	// --------------------------------------------------------

	session, err := s.sessionSvc.CreateSession(
		ctx,
		user.ID,
	)

	if err != nil {
		return database.User{}, database.Session{}, err
	}

	return user, session, nil
}

// ============================================================
// Session
// ============================================================

func (s *Service) Logout(
	ctx context.Context,
	sessionID pgtype.UUID,
) error {

	if !sessionID.Valid {
		return nil
	}

	return s.sessionSvc.DeleteSession(
		ctx,
		sessionID,
	)
}

func (s *Service) GetSession(
	ctx context.Context,
	sessionID pgtype.UUID,
) (database.Session, error) {

	return s.sessionSvc.GetSession(
		ctx,
		sessionID,
	)
}

func (s *Service) GetUserFromSession(
	ctx context.Context,
	sessionID pgtype.UUID,
) (database.User, error) {

	session, err := s.sessionSvc.GetSession(
		ctx,
		sessionID,
	)

	if err != nil {
		return database.User{}, err
	}

	user, err := s.userRepo.GetUserByID(
		ctx,
		session.UserID,
	)

	if err != nil {
		return database.User{},
			apperrors.UnauthorizedError(
				"invalid session",
			)
	}

	return user, nil
}

func (s *Service) LogoutAll(
	ctx context.Context,
	userID pgtype.UUID,
) error {

	return s.sessionSvc.DeleteUserSessions(
		ctx,
		userID,
	)
}

// ============================================================
// Username
// ============================================================

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
			return "",
				apperrors.InternalError(
					"failed to check username",
					err,
				)
		}

		if !exists {
			return candidate, nil
		}

		candidate = fmt.Sprintf(
			"%s%d",
			username,
			i+1,
		)
	}
}

// ============================================================
// Helpers
// ============================================================

func newUUID() pgtype.UUID {

	id := uuid.New()

	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}

func textFromString(value string) pgtype.Text {

	if value == "" {
		return pgtype.Text{
			Valid: false,
		}
	}

	return pgtype.Text{
		String: value,
		Valid:  true,
	}
}
