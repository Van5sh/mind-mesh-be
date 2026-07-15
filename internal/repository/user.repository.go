package repository

import (
	"context"

	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	q *database.Queries
}

func NewUserRepository(q *database.Queries) *UserRepository {
	return &UserRepository{
		q: q,
	}
}

func (r *UserRepository) CheckUsernameExists(
	ctx context.Context,
	username string,
) (bool, error) {
	return r.q.CheckUsernameExists(ctx, username)
}

func (r *UserRepository) CheckEmailExists(
	ctx context.Context,
	email string,
) (bool, error) {
	return r.q.CheckEmailExists(ctx, email)
}

// ---------- Create ----------

func (r *UserRepository) CreateUser(
	ctx context.Context,
	params database.CreateUserParams,
) (database.User, error) {
	return r.q.CreateUser(ctx, params)
}

func (r *UserRepository) CreateUserProfile(
	ctx context.Context,
	params database.CreateUserProfileParams,
) (database.UserProfile, error) {
	return r.q.CreateUserProfile(ctx, params)
}

func (r *UserRepository) GetUserByID(
	ctx context.Context,
	id pgtype.UUID,
) (database.User, error) {
	return r.q.GetUserByID(ctx, id)
}

func (r *UserRepository) GetUserByEmail(
	ctx context.Context,
	email string,
) (database.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

func (r *UserRepository) GetUserByUsername(
	ctx context.Context,
	username string,
) (database.User, error) {
	return r.q.GetUserByUsername(ctx, username)
}

func (r *UserRepository) GetUsersByIDs(
	ctx context.Context,
	ids []pgtype.UUID,
) ([]database.User, error) {
	return r.q.GetUsersByIDs(ctx, ids)
}

func (r *UserRepository) GetUserProfile(
	ctx context.Context,
	userID pgtype.UUID,
) (database.UserProfile, error) {
	return r.q.GetUserProfile(ctx, userID)
}

func (r *UserRepository) GetUserWithProfile(
	ctx context.Context,
	id pgtype.UUID,
) (database.GetUserWithProfileRow, error) {
	return r.q.GetUserWithProfile(ctx, id)
}

func (r *UserRepository) UpdateUser(
	ctx context.Context,
	params database.UpdateUserParams,
) (database.User, error) {
	return r.q.UpdateUser(ctx, params)
}

func (r *UserRepository) UpdateUserPassword(
	ctx context.Context,
	params database.UpdateUserPasswordParams,
) (database.User, error) {
	return r.q.UpdateUserPassword(ctx, params)
}

func (r *UserRepository) UpdateUserAvatar(
	ctx context.Context,
	params database.UpdateUserAvatarParams,
) (database.UserProfile, error) {
	return r.q.UpdateUserAvatar(ctx, params)
}

func (r *UserRepository) UpdateUserProfile(
	ctx context.Context,
	params database.UpdateUserProfileParams,
) (database.UserProfile, error) {
	return r.q.UpdateUserProfile(ctx, params)
}

func (r *UserRepository) DeleteUser(
	ctx context.Context,
	id pgtype.UUID,
) error {
	return r.q.DeleteUser(ctx, id)
}

func (r *UserRepository) DeleteUserProfile(
	ctx context.Context,
	userID pgtype.UUID,
) error {
	return r.q.DeleteUserProfile(ctx, userID)
}
