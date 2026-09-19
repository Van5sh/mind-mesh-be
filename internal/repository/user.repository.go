package repository

import (
	"context"
	"time"

	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
	q  *database.Queries
}

// CreateOAuthUser atomically provisions a user, profile, provider link, and
// initial session. A failed step rolls back every preceding write.
func (r *UserRepository) CreateOAuthUser(
	ctx context.Context,
	userParams database.CreateUserParams,
	provider string,
	providerUserID string,
	profileParams database.CreateUserProfileParams,
	sessionExpiresAt time.Time,
) (database.User, database.Session, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return database.User{}, database.Session{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := r.q.WithTx(tx)
	user, err := queries.CreateUser(ctx, userParams)
	if err != nil {
		return database.User{}, database.Session{}, err
	}

	profileParams.UserID = user.ID
	if _, err := queries.CreateUserProfile(ctx, profileParams); err != nil {
		return database.User{}, database.Session{}, err
	}

	if _, err := queries.CreateOAuthAccount(ctx, database.CreateOAuthAccountParams{
		UserID:         user.ID,
		Provider:       provider,
		ProviderUserID: providerUserID,
	}); err != nil {
		return database.User{}, database.Session{}, err
	}

	session, err := queries.CreateSession(ctx, database.CreateSessionParams{
		UserID: user.ID,
		ExpiresAt: pgtype.Timestamptz{
			Time:  sessionExpiresAt,
			Valid: true,
		},
	})
	if err != nil {
		return database.User{}, database.Session{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return database.User{}, database.Session{}, err
	}

	return user, session, nil
}

func NewUserRepository(db *pgxpool.Pool, q *database.Queries) *UserRepository {
	return &UserRepository{
		db: db,
		q:  q,
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

func (r *UserRepository) CreateUser(
	ctx context.Context,
	params database.CreateUserParams,
) (database.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return database.User{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := r.q.WithTx(tx)
	user, err := queries.CreateUser(ctx, params)
	if err != nil {
		return database.User{}, err
	}

	// Registration does not collect a person's name. Use the username as the
	// initial profile name; it can be replaced with UpdateUserProfile after
	// registration.
	if _, err := queries.CreateUserProfile(ctx, database.CreateUserProfileParams{
		UserID:    user.ID,
		FirstName: user.Username,
		LastName:  user.Username,
	}); err != nil {
		return database.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return database.User{}, err
	}

	return user, nil
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

func (r *UserRepository) GetAllUsers(ctx context.Context) ([]database.User, error) {
	return r.q.GetAllUsers(ctx)
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

func (r *UserRepository) GetUserProfilesByUserIDs(ctx context.Context, userIDs []pgtype.UUID) ([]database.UserProfile, error) {
	return r.q.GetUserProfilesByUserIDs(ctx, userIDs)
}
