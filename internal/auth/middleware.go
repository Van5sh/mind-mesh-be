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

func Middleware(
	service *Service,
) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			sessionID, err := parseSessionID(cookie.Value)
			if err != nil {
				ClearSessionCookie(w)
				next.ServeHTTP(w, r)
				return
			}

			user, err := service.GetUserFromSession(
				r.Context(),
				sessionID,
			)
			if err != nil {
				ClearSessionCookie(w)
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				userContextKey,
				user,
			)

			ctx = context.WithValue(
				ctx,
				sessionContextKey,
				sessionID,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		})
	}
}

// ============================================================
// Session ID Parsing
// ============================================================

func parseSessionID(value string) (pgtype.UUID, error) {
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
// Context Helpers
// ============================================================

// UserFromContext returns the authenticated user stored in ctx.
func UserFromContext(ctx context.Context) (database.User, bool) {
	user, ok := ctx.Value(userContextKey).(database.User)
	return user, ok
}

// UserIDFromContext returns the authenticated user's ID.
func UserIDFromContext(ctx context.Context) (pgtype.UUID, bool) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return pgtype.UUID{}, false
	}

	return user.ID, true
}

// SessionFromContext returns the authenticated session ID.
func SessionFromContext(ctx context.Context) (pgtype.UUID, bool) {
	session, ok := ctx.Value(sessionContextKey).(pgtype.UUID)
	if !ok || !session.Valid {
		return pgtype.UUID{}, false
	}

	return session, true
}

// SessionIDFromContext is an alias-style helper for retrieving
// the current session ID.
func SessionIDFromContext(ctx context.Context) (pgtype.UUID, bool) {
	return SessionFromContext(ctx)
}

// ============================================================
// Required Authentication
// ============================================================

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		_, ok := UserFromContext(r.Context())

		if !ok {
			http.Error(
				w,
				"authentication required",
				http.StatusUnauthorized,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
