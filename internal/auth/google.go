package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleProvider struct {
	Config *oauth2.Config
}

type GoogleUser struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

func NewGoogleProvider(
	clientID string,
	clientSecret string,
	redirectURL string,
) *GoogleProvider {
	return &GoogleProvider{
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,

			Endpoint: google.Endpoint,

			Scopes: []string{
				"openid",
				"profile",
				"email",
			},
		},
	}
}

// AuthURL creates the URL where the user is redirected
// to authenticate with Google.
func (p *GoogleProvider) AuthURL(state string) string {
	return p.Config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
	)
}

// Exchange exchanges the authorization code for a Google
// access token and retrieves the user's identity.
func (p *GoogleProvider) Exchange(
	ctx context.Context,
	code string,
) (*GoogleUser, error) {

	token, err := p.Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to exchange google authorization code: %w",
			err,
		)
	}

	client := p.Config.Client(ctx, token)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://openidconnect.googleapis.com/v1/userinfo",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create google userinfo request: %w",
			err,
		)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to request google userinfo: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"google userinfo returned status %d",
			resp.StatusCode,
		)
	}

	var user GoogleUser

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf(
			"failed to decode google userinfo: %w",
			err,
		)
	}

	if user.Sub == "" {
		return nil, fmt.Errorf(
			"google user does not have a subject",
		)
	}

	if user.Email == "" {
		return nil, fmt.Errorf(
			"google user does not have an email",
		)
	}

	if !user.EmailVerified {
		return nil, fmt.Errorf(
			"google email is not verified",
		)
	}

	return &user, nil
}
