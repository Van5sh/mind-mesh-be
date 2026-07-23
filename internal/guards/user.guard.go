package guards

import (
	"context"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

type UserGuard struct {
	repo *repository.UserRepository
}

func NewUserGuard(repo *repository.UserRepository) *UserGuard {
	return &UserGuard{repo: repo}
}

func (g *UserGuard) EnsureUserExists(ctx context.Context, userID pgtype.UUID) (database.User, error) {
	user, err := g.repo.GetUserByID(ctx, userID)
	if err != nil {
		if isNoRows(err) {
			return database.User{}, apperrors.NotFoundError("user not found")
		}
		return database.User{}, apperrors.InternalError("failed to fetch user", err)
	}
	return user, nil
}

func (g *UserGuard) EnsureUserProfileExists(ctx context.Context, userID pgtype.UUID) (database.UserProfile, error) {
	profile, err := g.repo.GetUserProfile(ctx, userID)
	if err != nil {
		if isNoRows(err) {
			return database.UserProfile{}, apperrors.NotFoundError("user profile not found")
		}
		return database.UserProfile{}, apperrors.InternalError("failed to fetch user profile", err)
	}
	return profile, nil
}

func (g *UserGuard) EnsureUsernameAvailable(ctx context.Context, username string) error {
	exists, err := g.repo.CheckUsernameExists(ctx, username)
	if err != nil {
		return apperrors.InternalError("failed to check username", err)
	}
	if exists {
		return apperrors.ConflictError("username already exists")
	}
	return nil
}

func (g *UserGuard) EnsureEmailAvailable(ctx context.Context, email string) error {
	exists, err := g.repo.CheckEmailExists(ctx, email)
	if err != nil {
		return apperrors.InternalError("failed to check email", err)
	}
	if exists {
		return apperrors.ConflictError("email already exists")
	}
	return nil
}
