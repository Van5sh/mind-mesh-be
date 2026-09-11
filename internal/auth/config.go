package auth

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"example/hello/internal/repository"
	"example/hello/internal/services"
)

const (
	defaultFrontendURL = "http://localhost:3000"
	googleCallbackPath = "/auth/google/callback"
	githubCallbackPath = "/auth/github/callback"
)

// NewOAuthHandlerFromEnvironment builds the complete OAuth flow from the
// application's environment. Register the returned handler's four HTTP
// methods at the callback URLs configured with Google and GitHub.
func NewOAuthHandlerFromEnvironment(
	userRepo *repository.UserRepository,
	oauthRepo *repository.OAuthRepository,
	sessionSvc *services.SessionService,
) (*OAuthHandler, error) {
	baseURL := strings.TrimRight(os.Getenv("API_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	googleClientID, err := requiredEnv("GOOGLE_CLIENT_ID")
	if err != nil {
		return nil, err
	}
	googleClientSecret, err := requiredEnv("GOOGLE_CLIENT_SECRET")
	if err != nil {
		return nil, err
	}
	githubClientID, err := requiredEnv("GITHUB_CLIENT_ID")
	if err != nil {
		return nil, err
	}
	githubClientSecret, err := requiredEnv("GITHUB_CLIENT_SECRET")
	if err != nil {
		return nil, err
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

	google := NewGoogleProvider(googleClientID, googleClientSecret, baseURL+googleCallbackPath)
	github := NewGitHubProvider(githubClientID, githubClientSecret, baseURL+githubCallbackPath)
	service := NewService(userRepo, oauthRepo, sessionSvc, google, github)

	return NewOAuthHandler(service, google, github, frontendURL, secureCookie, 7*24*time.Hour), nil
}

func requiredEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%s is required for OAuth", name)
	}
	return value, nil
}
