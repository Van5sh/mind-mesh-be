package auth

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"example/hello/internal/repository"
	"example/hello/internal/services"
)

const defaultFrontendURL = "http://localhost:3000"

// NewOAuthHandlerFromEnvironment builds authentication from the
// application's environment: a Firebase Admin SDK client (verifies ID
// tokens for Google/GitHub sign-ins that happened client-side via the
// Firebase JS SDK) plus the existing session-cookie machinery. Register
// the returned handler's FirebaseLogin method at POST /auth/firebase.
func NewOAuthHandlerFromEnvironment(
	ctx context.Context,
	userRepo *repository.UserRepository,
	oauthRepo *repository.OAuthRepository,
	sessionSvc *services.SessionService,
) (*OAuthHandler, error) {
	credentialsPath, err := requiredEnv("FIREBASE_CREDENTIALS_PATH")
	if err != nil {
		return nil, err
	}

	firebaseProvider, err := NewFirebaseProvider(ctx, credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("initialize firebase: %w", err)
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = defaultFrontendURL
	}

	secureCookie := os.Getenv("APP_ENV") == "production"
	if value := os.Getenv("AUTH_COOKIE_SECURE"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("AUTH_COOKIE_SECURE must be true or false: %w", err)
		}
		secureCookie = parsed
	}

	service := NewService(userRepo, oauthRepo, sessionSvc, firebaseProvider)

	return NewOAuthHandler(service, frontendURL, secureCookie, 7*24*time.Hour), nil
}

func requiredEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%s is required for authentication", name)
	}
	return value, nil
}
