package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"golang.org/x/oauth2"
)

type GitHubProvider struct {
	Config *oauth2.Config
}

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func NewGitHubProvider(
	clientID string,
	clientSecret string,
	redirectURL string,
) *GitHubProvider {
	return &GitHubProvider{
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,

			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},

			Scopes: []string{
				"read:user",
				"user:email",
			},
		},
	}
}

// AuthURL creates the URL where the user is redirected
// to authenticate with GitHub.
func (p *GitHubProvider) AuthURL(state string) string {
	return p.Config.AuthCodeURL(state)
}

// Exchange exchanges the authorization code for a GitHub
// access token and retrieves the user's identity.
func (p *GitHubProvider) Exchange(
	ctx context.Context,
	code string,
) (*GitHubUser, error) {

	token, err := p.Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to exchange github authorization code: %w",
			err,
		)
	}

	client := p.Config.Client(ctx, token)

	user, err := getGitHubUser(ctx, client)
	if err != nil {
		return nil, err
	}

	// GitHub may not return an email in /user.
	// Retrieve the user's verified primary email.
	if user.Email == "" {
		email, err := getGitHubPrimaryEmail(ctx, client)
		if err != nil {
			return nil, err
		}

		user.Email = email
	}

	if user.Email == "" {
		return nil, fmt.Errorf(
			"github user does not have a usable email",
		)
	}

	return user, nil
}

func getGitHubUser(
	ctx context.Context,
	client *http.Client,
) (*GitHubUser, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.github.com/user",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create github user request: %w",
			err,
		)
	}

	req.Header.Set(
		"Accept",
		"application/vnd.github+json",
	)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to request github user: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"github user endpoint returned status %d",
			resp.StatusCode,
		)
	}

	var user GitHubUser

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf(
			"failed to decode github user: %w",
			err,
		)
	}

	if user.ID == 0 {
		return nil, fmt.Errorf(
			"github user does not have an id",
		)
	}

	return &user, nil
}

func getGitHubPrimaryEmail(
	ctx context.Context,
	client *http.Client,
) (string, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.github.com/user/emails",
		nil,
	)
	if err != nil {
		return "", fmt.Errorf(
			"failed to create github email request: %w",
			err,
		)
	}

	req.Header.Set(
		"Accept",
		"application/vnd.github+json",
	)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf(
			"failed to request github emails: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"github email endpoint returned status %d",
			resp.StatusCode,
		)
	}

	var emails []GitHubEmail

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf(
			"failed to decode github emails: %w",
			err,
		)
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	return "", nil
}

// GitHubUserID returns the GitHub provider ID as a string
// suitable for oauth_accounts.provider_user_id.
func GitHubUserID(user *GitHubUser) string {
	return strconv.FormatInt(user.ID, 10)
}
