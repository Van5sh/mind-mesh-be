package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"time"

	"example/hello/internal/database"

	"github.com/google/uuid"
)

const (
	oauthStateCookieName = "oauth_state"
	sessionCookieName    = "session_id"

	oauthStateMaxAge = 10 * 60 // 10 minutes
)

type OAuthHandler struct {
	service *Service
	google  *GoogleProvider
	github  *GitHubProvider

	frontendURL      string
	secureCookie     bool
	sessionCookieAge time.Duration
}

func NewOAuthHandler(
	service *Service,
	google *GoogleProvider,
	github *GitHubProvider,
	frontendURL string,
	secureCookie bool,
	sessionCookieAge time.Duration,
) *OAuthHandler {
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	if sessionCookieAge <= 0 {
		sessionCookieAge = 7 * 24 * time.Hour
	}

	return &OAuthHandler{
		service:          service,
		google:           google,
		github:           github,
		frontendURL:      frontendURL,
		secureCookie:     secureCookie,
		sessionCookieAge: sessionCookieAge,
	}
}

// Middleware attaches the authenticated user and session to request contexts.
// Wrap API handlers that need to read the login session with it.
func (h *OAuthHandler) Middleware(next http.Handler) http.Handler {
	return Middleware(h.service)(next)
}

// ============================================================
// Google
// ============================================================

// GoogleLogin redirects the user to Google's OAuth page.
func (h *OAuthHandler) GoogleLogin(
	w http.ResponseWriter,
	r *http.Request,
) {
	state, err := generateOAuthState()
	if err != nil {
		http.Error(
			w,
			"failed to generate oauth state",
			http.StatusInternalServerError,
		)
		return
	}

	h.setOAuthStateCookie(w, state)

	http.Redirect(
		w,
		r,
		h.google.AuthURL(state),
		http.StatusTemporaryRedirect,
	)
}

// GoogleCallback handles Google's OAuth callback.
func (h *OAuthHandler) GoogleCallback(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.URL.Query().Get("error") != "" {
		http.Error(
			w,
			"google authentication was cancelled or denied",
			http.StatusUnauthorized,
		)
		return
	}

	state := r.URL.Query().Get("state")

	if !validateOAuthState(r, state) {
		http.Error(
			w,
			"invalid oauth state",
			http.StatusBadRequest,
		)
		return
	}
	h.clearOAuthStateCookie(w)

	code := r.URL.Query().Get("code")

	if code == "" {
		http.Error(
			w,
			"missing authorization code",
			http.StatusBadRequest,
		)
		return
	}

	_, session, err := h.service.LoginWithGoogle(
		r.Context(),
		code,
	)
	if err != nil {
		http.Error(
			w,
			"google authentication failed",
			http.StatusUnauthorized,
		)
		return
	}

	h.setSessionCookie(w, session)

	http.Redirect(
		w,
		r,
		h.frontendURL,
		http.StatusSeeOther,
	)
}

// ============================================================
// GitHub
// ============================================================

// GitHubLogin redirects the user to GitHub's OAuth page.
func (h *OAuthHandler) GitHubLogin(
	w http.ResponseWriter,
	r *http.Request,
) {
	state, err := generateOAuthState()
	if err != nil {
		http.Error(
			w,
			"failed to generate oauth state",
			http.StatusInternalServerError,
		)
		return
	}

	h.setOAuthStateCookie(w, state)

	http.Redirect(
		w,
		r,
		h.github.AuthURL(state),
		http.StatusTemporaryRedirect,
	)
}

// GitHubCallback handles GitHub's OAuth callback.
func (h *OAuthHandler) GitHubCallback(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.URL.Query().Get("error") != "" {
		http.Error(
			w,
			"github authentication was cancelled or denied",
			http.StatusUnauthorized,
		)
		return
	}

	state := r.URL.Query().Get("state")

	if !validateOAuthState(r, state) {
		http.Error(
			w,
			"invalid oauth state",
			http.StatusBadRequest,
		)
		return
	}
	h.clearOAuthStateCookie(w)

	code := r.URL.Query().Get("code")

	if code == "" {
		http.Error(
			w,
			"missing authorization code",
			http.StatusBadRequest,
		)
		return
	}

	_, session, err := h.service.LoginWithGitHub(
		r.Context(),
		code,
	)
	if err != nil {
		http.Error(
			w,
			"github authentication failed",
			http.StatusUnauthorized,
		)
		return
	}

	h.setSessionCookie(w, session)

	http.Redirect(
		w,
		r,
		h.frontendURL,
		http.StatusSeeOther,
	)
}

// ============================================================
// OAuth State
// ============================================================

// generateOAuthState creates a cryptographically secure
// random state value used to protect the OAuth flow from CSRF.
func generateOAuthState() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// setOAuthStateCookie stores the OAuth state in an HTTP-only cookie.
func (h *OAuthHandler) setOAuthStateCookie(
	w http.ResponseWriter,
	state string,
) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,

		Secure: h.secureCookie,

		SameSite: http.SameSiteLaxMode,

		MaxAge: oauthStateMaxAge,
	})
}

// validateOAuthState validates the state returned by the provider.
func validateOAuthState(
	r *http.Request,
	state string,
) bool {
	if state == "" {
		return false
	}

	cookie, err := r.Cookie(oauthStateCookieName)
	if err != nil {
		return false
	}

	if cookie.Value == "" {
		return false
	}

	if len(cookie.Value) != len(state) {
		return false
	}

	return subtle.ConstantTimeCompare(
		[]byte(cookie.Value),
		[]byte(state),
	) == 1
}

// clearOAuthStateCookie removes the OAuth state cookie.
func (h *OAuthHandler) clearOAuthStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// ============================================================
// Session Cookie
// ============================================================

// setSessionCookie stores the authenticated session ID
// in an HTTP-only cookie.
func (h *OAuthHandler) setSessionCookie(
	w http.ResponseWriter,
	session database.Session,
) {
	sessionID := uuid.UUID(session.ID.Bytes).String()

	http.SetCookie(w, &http.Cookie{
		Name:  sessionCookieName,
		Value: sessionID,
		Path:  "/",

		HttpOnly: true,

		Secure: h.secureCookie,

		SameSite: http.SameSiteLaxMode,

		Expires: time.Now().Add(h.sessionCookieAge),
		MaxAge:  int(h.sessionCookieAge.Seconds()),
	})
}

// ClearSessionCookie removes the current session cookie.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
