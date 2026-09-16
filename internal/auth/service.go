package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/repository"
	"example/hello/internal/services"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	userRepo   *repository.UserRepository
	oauthRepo  *repository.OAuthRepository
	sessionSvc *services.SessionService

	firebase *FirebaseProvider
}

func NewService(
	userRepo *repository.UserRepository,
	oauthRepo *repository.OAuthRepository,
	sessionSvc *services.SessionService,
	firebase *FirebaseProvider,
) *Service {
	return &Service{
		userRepo:   userRepo,
		oauthRepo:  oauthRepo,
		sessionSvc: sessionSvc,
		firebase:   firebase,
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

// LoginWithFirebase verifies a Firebase ID token - obtained client-side via
// the Firebase JS SDK after the user completes a Google or GitHub sign-in -
// and finds-or-creates the matching local user, exactly like the old
// direct-OAuth2 flow did. This is the only entry point into
// LoginWithOAuth now; Firebase's console is configured with both Google
// and GitHub as sign-in providers, so both arrive here identically.
func (s *Service) LoginWithFirebase(
	ctx context.Context,
	idToken string,
) (database.User, database.Session, error) {

	identity, err := s.firebase.VerifyIDToken(ctx, idToken)
	if err != nil {
		return database.User{}, database.Session{},
			apperrors.UnauthorizedError("invalid firebase id token")
	}

	firstName, lastName := splitName(identity.Name)

	username := firstName
	if username == "" {
		username = emailLocalPart(identity.Email)
	}

	oauthUser := OAuthUser{
		Provider:       identity.Provider,
		ProviderUserID: identity.ProviderUserID,
		Email:          identity.Email,
		Username:       username,
		FirstName:      firstName,
		LastName:       lastName,
		AvatarURL:      identity.AvatarURL,
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
	if !errors.Is(err, pgx.ErrNoRows) {
		return database.User{}, database.Session{}, apperrors.InternalError(
			"failed to look up OAuth account",
			err,
		)
	}

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
	if !errors.Is(err, pgx.ErrNoRows) {
		return database.User{}, database.Session{}, apperrors.InternalError(
			"failed to look up user by OAuth email",
			err,
		)
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

	user, session, err := s.userRepo.CreateOAuthUser(
		ctx,
		database.CreateUserParams{
			ID:       userID,
			Username: username,
			Email:    oauthUser.Email,
		},
		oauthUser.Provider,
		oauthUser.ProviderUserID,
		database.CreateUserProfileParams{
			FirstName: oauthUser.FirstName,
			LastName:  oauthUser.LastName,
			Bio:       textFromString(""),
			AvatarUrl: textFromString(oauthUser.AvatarURL),
		},
		time.Now().Add(7*24*time.Hour),
	)

	if err != nil {
		return database.User{}, database.Session{},
			apperrors.InternalError(
				"failed to create OAuth user",
				err,
			)
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

// splitName splits a Firebase display name ("Ada Lovelace") into a first
// and last name the way the old Google flow's given_name/family_name
// claims used to arrive pre-split. Firebase only gives one "name" claim,
// so this is a best-effort heuristic, not a guarantee.
func splitName(name string) (first string, last string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ""
	}

	parts := strings.SplitN(name, " ", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// emailLocalPart returns the part of an email before "@", used as a
// username seed when there's no display name to derive one from.
func emailLocalPart(email string) string {
	at := strings.Index(email, "@")
	if at <= 0 {
		return email
	}
	return email[:at]
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
