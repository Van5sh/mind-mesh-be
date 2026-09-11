package auth

import (
	"context"
	"net/http"

	"example/hello/internal/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type contextKey string

const (
	userContextKey    contextKey = "auth_user"
	sessionContextKey contextKey = "auth_session"
)

// ============================================================
// Authentication Middleware
// ============================================================

// Middleware reads the session_id cookie, validates the session,
// loads the authenticated user, and stores both in the request
// context.
//
// Anonymous requests are allowed through.
// Use RequireAuth when authentication is mandatory.
func Middleware(
	service *Service,
) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {

				// ------------------------------------------------
				// Get session cookie
				// ------------------------------------------------

				cookie, err := r.Cookie(sessionCookieName)

				if err != nil {
					// No session cookie.
					// User is anonymous.
					next.ServeHTTP(w, r)
					return
				}

				// ------------------------------------------------
				// Parse session ID
				// ------------------------------------------------

				sessionID, err := parseSessionID(cookie.Value)

				if err != nil {
					// Invalid cookie.
					ClearSessionCookie(w)

					next.ServeHTTP(w, r)
					return
				}

				// ------------------------------------------------
				// Load session + user
				// ------------------------------------------------
	
				user, err :=
					service.GetUserFromSession(
						r.Context(),
						sessionID,
					)
	
				if err != nil {
					// Session is invalid/expired.
					ClearSessionCookie(w)
	
					next.ServeHTTP(w, r)
					return
				}
	
				// ------------------------------------------------
				// Store user in context
				// ------------------------------------------------
	
				ctx := context.WithValue(
					r.Context(),
					userContextKey,
					user,
				)
	
				// ------------------------------------------------
				// Store session in context
				// ------------------------------------------------
	
				ctx = context.WithValue(
					ctx,
					sessionContextKey,
					&pgtype.UUID{
						Bytes: sessionID.Bytes,
						Valid: sessionID.Valid,
					},
				)

				// ------------------------------------------------
				// Continue request
				// ------------------------------------------------

				next.ServeHTTP(
					w,
					r.WithContext(ctx),
				)
			},
		)
	}
}

// ============================================================
// Session ID
// ============================================================

func parseSessionID(
	value string,
) (pgtype.UUID, error) {

	id, err := uuid.Parse(value)

	if err != nil {
		return pgtype.UUID{}, err
	}

	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}, nil
}

// ============================================================
// User Context
// ============================================================

// UserFromContext returns the authenticated user.
//
// The second return value is false when the request is anonymous.
func UserFromContext(
	ctx context.Context,
) (database.User, bool) {

	user, ok := ctx.Value(
		userContextKey,
	).(database.User)

	return user, ok
}

// ============================================================
// Session Context
// ============================================================

// SessionFromContext returns the current session.
//
// The second return value is false when the request is anonymous.
func SessionFromContext(
	ctx context.Context,
) (database.Session, bool) {

	session, ok := ctx.Value(
		sessionContextKey,
	).(database.Session)

	return session, ok
}

// ============================================================
// User ID
// ============================================================

// UserIDFromContext returns the authenticated user's ID.
func UserIDFromContext(
	ctx context.Context,
) (pgtype.UUID, bool) {

	user, ok := UserFromContext(ctx)

	if !ok {
		return pgtype.UUID{}, false
	}

	return user.ID, true
}

// ============================================================
// Session ID
// ============================================================

// SessionIDFromContext returns the current session ID.
func SessionIDFromContext(
	ctx context.Context,
) (pgtype.UUID, bool) {

	session, ok := SessionFromContext(ctx)

	if !ok {
		return pgtype.UUID{}, false
	}

	return session.ID, true
}

// ============================================================
// Require Authentication
// ============================================================

// RequireAuth rejects anonymous requests.
//
// Use this for routes/resolvers that require a logged-in user.
func RequireAuth(
	next http.Handler,
) http.Handler {

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			_, ok := UserFromContext(
				r.Context(),
			)

			if !ok {
				http.Error(
					w,
					"authentication required",
					http.StatusUnauthorized,
				)

				return
			}

			next.ServeHTTP(w, r)
		},
	)
}
