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

				user, err :=
					service.GetUserFromSession(
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
					&pgtype.UUID{
						Bytes: sessionID.Bytes,
						Valid: sessionID.Valid,
					},
				)
				next.ServeHTTP(
					w,
					r.WithContext(ctx),
				)
			},
		)
	}
}

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

func UserFromContext(
	ctx context.Context,
) (database.User, bool) {

	user, ok := ctx.Value(
		userContextKey,
	).(database.User)

	return user, ok
}


func SessionFromContext(
	ctx context.Context,
) (pgtype.UUID, bool) {

	session, ok := ctx.Value(
		sessionContextKey,
	).(*pgtype.UUID)

	if !ok || session == nil {
		return pgtype.UUID{}, false
	}

	return *session, true
}

func UserIDFromContext(
	ctx context.Context,
) (pgtype.UUID, bool) {

	user, ok := UserFromContext(ctx)

	if !ok {
		return pgtype.UUID{}, false
	}

	return user.ID, true
}

func SessionIDFromContext(
	ctx context.Context,
) (pgtype.UUID, bool) {

	session, ok := SessionFromContext(ctx)

	if !ok {
		return pgtype.UUID{}, false
	}

	return session, true
}

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
