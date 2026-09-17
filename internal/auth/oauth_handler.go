package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"example/hello/internal/database"

	"github.com/google/uuid"
)

const sessionCookieName = "session_id"

// OAuthHandler exposes the one HTTP entry point authentication needs
// outside of GraphQL: establishing a session from a verified Firebase ID
// token. (Named OAuthHandler for continuity with the rest of the codebase
// - it's still "the thing that turns a Google/GitHub sign-in into a
// session," just via Firebase now instead of running the OAuth2 dance
// itself.)
type OAuthHandler struct {
	service *Service

	frontendURL      string
	secureCookie     bool
	sessionCookieAge time.Duration
}

func NewOAuthHandler(
	service *Service,
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
// Firebase
// ============================================================

type firebaseLoginRequest struct {
	IDToken string `json:"idToken"`
}

type firebaseLoginResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// FirebaseLogin verifies a Firebase ID token - the frontend gets this by
// running the actual Google/GitHub sign-in through the Firebase JS SDK,
// entirely client-side - and, if valid, establishes the same session-cookie
// session every other part of this API already relies on. Unlike the old
// OAuth2 redirect flow, this is a plain JSON POST from frontend JavaScript,
// not a browser navigation, so it responds with JSON instead of a redirect.
func (h *OAuthHandler) FirebaseLogin(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body firebaseLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.IDToken == "" {
		http.Error(w, "missing idToken", http.StatusBadRequest)
		return
	}

	user, session, err := h.service.LoginWithFirebase(r.Context(), body.IDToken)
	if err != nil {
		http.Error(w, "firebase authentication failed", http.StatusUnauthorized)
		return
	}

	h.setSessionCookie(w, session)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(firebaseLoginResponse{
		ID:       uuid.UUID(user.ID.Bytes).String(),
		Username: user.Username,
		Email:    user.Email,
	})
}

// ============================================================
// Logout
// ============================================================

// Logout invalidates the current session server-side (deletes the row, so
// the cookie can't be replayed even if someone captured it) and clears the
// cookie. A missing or already-invalid cookie is not an error - logging
// out a session that's already gone still counts as success, matching how
// logout endpoints are generally expected to behave (idempotent, never
// blocks the client from ending up logged out).
func (h *OAuthHandler) Logout(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		if sessionID, err := parseSessionID(cookie.Value); err == nil {
			if err := h.service.Logout(r.Context(), sessionID); err != nil {
				// The cookie gets cleared regardless (below) - a failed
				// server-side delete shouldn't leave the client stuck
				// thinking it's still logged in.
				log.Printf("logout: failed to delete session: %v", err)
			}
		}
	}

	ClearSessionCookie(w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
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
