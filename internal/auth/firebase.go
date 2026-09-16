package auth

import (
	"context"
	"fmt"
	"strings"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// FirebaseProvider verifies ID tokens issued by Firebase Authentication.
// Google and GitHub sign-in are both configured as providers *inside*
// Firebase (Firebase console, not this backend) - the frontend uses the
// Firebase JS SDK to run the actual OAuth handshake with whichever
// provider the user picks, and this backend never talks to Google or
// GitHub directly anymore. Its only job is verifying the ID token Firebase
// hands the frontend afterward.
type FirebaseProvider struct {
	client *firebaseauth.Client
}

// NewFirebaseProvider initializes the Firebase Admin SDK from a service
// account credentials JSON file (FIREBASE_CREDENTIALS_PATH).
func NewFirebaseProvider(ctx context.Context, credentialsPath string) (*FirebaseProvider, error) {
	app, err := firebase.NewApp(
		ctx,
		nil,
		option.WithCredentialsFile(credentialsPath),
	)
	if err != nil {
		return nil, fmt.Errorf("initialize firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize firebase auth client: %w", err)
	}

	return &FirebaseProvider{client: client}, nil
}

// VerifiedIdentity is everything needed to find-or-create a local user out
// of a verified Firebase ID token - deliberately the same shape regardless
// of which underlying provider the user actually signed in with.
type VerifiedIdentity struct {
	// Provider is "GOOGLE" or "GITHUB" - kept in the same form the old
	// direct-OAuth2 flow used (oauth_accounts.provider), so existing rows
	// and the account-linking logic in Service.LoginWithOAuth don't change.
	Provider       string
	ProviderUserID string
	Email          string
	Name           string
	AvatarURL      string
}

// VerifyIDToken verifies a Firebase ID token (obtained client-side via the
// Firebase JS SDK) and extracts the identity Service.LoginWithFirebase
// needs. Returns an error if the token is invalid/expired, or if it wasn't
// issued for a sign-in provider this backend supports.
func (p *FirebaseProvider) VerifyIDToken(ctx context.Context, idToken string) (VerifiedIdentity, error) {
	token, err := p.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return VerifiedIdentity{}, fmt.Errorf("verify firebase id token: %w", err)
	}

	signInProvider := firebaseSignInProvider(token)

	provider, err := normalizeSignInProvider(signInProvider)
	if err != nil {
		return VerifiedIdentity{}, err
	}

	email, _ := token.Claims["email"].(string)
	name, _ := token.Claims["name"].(string)
	picture, _ := token.Claims["picture"].(string)

	return VerifiedIdentity{
		Provider:       provider,
		ProviderUserID: firebaseProviderUserID(token, signInProvider),
		Email:          email,
		Name:           name,
		AvatarURL:      picture,
	}, nil
}

// firebaseSignInProvider reads Firebase's own "which provider was this
// token issued from" claim: token.Claims["firebase"]["sign_in_provider"],
// e.g. "google.com" or "github.com".
func firebaseSignInProvider(token *firebaseauth.Token) string {
	firebaseClaims, ok := token.Claims["firebase"].(map[string]interface{})
	if !ok {
		return ""
	}

	signInProvider, _ := firebaseClaims["sign_in_provider"].(string)
	return signInProvider
}

// firebaseProviderUserID prefers the raw per-provider identity Firebase
// recorded (token.Claims["firebase"]["identities"][signInProvider][0]) -
// e.g. Google's "sub" or GitHub's numeric ID - falling back to Firebase's
// own UID if that shape isn't present for some reason. Either is a stable,
// unique identifier; preferring the provider-native one keeps values
// consistent with what the old direct-OAuth2 flow used to store.
func firebaseProviderUserID(token *firebaseauth.Token, signInProvider string) string {
	firebaseClaims, ok := token.Claims["firebase"].(map[string]interface{})
	if ok {
		identities, ok := firebaseClaims["identities"].(map[string]interface{})
		if ok {
			if ids, ok := identities[signInProvider].([]interface{}); ok && len(ids) > 0 {
				if id, ok := ids[0].(string); ok && id != "" {
					return id
				}
			}
		}
	}

	return token.UID
}

// normalizeSignInProvider maps Firebase's provider identifier to the
// GOOGLE/GITHUB values already used throughout this codebase
// (oauth_accounts.provider, Service.OAuthUser.Provider).
func normalizeSignInProvider(signInProvider string) (string, error) {
	switch strings.ToLower(signInProvider) {
	case "google.com":
		return "GOOGLE", nil
	case "github.com":
		return "GITHUB", nil
	default:
		return "", fmt.Errorf("unsupported firebase sign-in provider: %q", signInProvider)
	}
}
