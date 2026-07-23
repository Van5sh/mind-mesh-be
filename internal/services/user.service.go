package services

import (
	"context"
	"errors"
	"fmt"

	"example/hello/internal/guards"
	"example/hello/internal/database"
	"example/hello/internal/repository"
	"example/hello/internal/validators"

	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUsernameExists  = errors.New("username already exists")
	ErrEmailExists     = errors.New("email already exists")
	ErrProfileNotFound = errors.New("user profile not found")
)

type UserService struct {
	repo  *repository.UserRepository
	guard *guards.UserGuard
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo:  repo,
		guard: guards.NewUserGuard(repo),
	}
}

func (s *UserService) CreateUser(
	ctx context.Context,
	params database.CreateUserParams,
) (database.User, error) {
	if err := validators.ValidateUsername(params.Username); err != nil {
		return database.User{}, err
	}
	if err := validators.ValidateEmail(params.Email); err != nil {
		return database.User{}, err
	}
	if err := validators.ValidatePasswordHash(params.PasswordHash); err != nil {
		return database.User{}, err
	}

	if err := s.guard.EnsureUsernameAvailable(ctx, params.Username); err != nil {
		if errors.Is(err, ErrUsernameExists) {
			return database.User{}, err
		}
		return database.User{}, err
	}

	if err := s.guard.EnsureEmailAvailable(ctx, params.Email); err != nil {
		return database.User{}, err
	}

	user, err := s.repo.CreateUser(ctx, params)
	if err != nil {
		return database.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *UserService) CreateUserProfile(
	ctx context.Context,
	params database.CreateUserProfileParams,
) (database.UserProfile, error) {
	if err := validators.ValidateFirstName(params.FirstName); err != nil {
		return database.UserProfile{}, err
	}
	if err := validators.ValidateLastName(params.LastName); err != nil {
		return database.UserProfile{}, err
	}
	if err := validators.ValidateUserBio(params.Bio.String); err != nil {
		return database.UserProfile{}, err
	}
	if err := validators.ValidateAvatarURL(params.AvatarUrl.String); err != nil {
		return database.UserProfile{}, err
	}

	_, err := s.guard.EnsureUserExists(ctx, params.UserID)
	if err != nil {
		return database.UserProfile{}, err
	}

	profile, err := s.repo.CreateUserProfile(ctx, params)
	if err != nil {
		return database.UserProfile{}, fmt.Errorf("create user profile: %w", err)
	}

	return profile, nil
}

func (s *UserService) GetUserByID(
	ctx context.Context,
	id pgtype.UUID,
) (database.User, error) {

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return database.User{}, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUserByEmail(
	ctx context.Context,
	email string,
) (database.User, error) {

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return database.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUserByUsername(
	ctx context.Context,
	username string,
) (database.User, error) {

	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return database.User{}, fmt.Errorf("get user by username: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUsersByIDs(
	ctx context.Context,
	ids []pgtype.UUID,
) ([]database.User, error) {

	if len(ids) == 0 {
		return []database.User{}, nil
	}

	users, err := s.repo.GetUsersByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("get users by ids: %w", err)
	}

	return users, nil
}

func (s *UserService) GetUserProfile(
	ctx context.Context,
	userID pgtype.UUID,
) (database.UserProfile, error) {

	profile, err := s.repo.GetUserProfile(ctx, userID)
	if err != nil {
		return database.UserProfile{}, fmt.Errorf("get user profile: %w", err)
	}

	return profile, nil
}

func (s *UserService) GetUserWithProfile(
	ctx context.Context,
	userID pgtype.UUID,
) (database.GetUserWithProfileRow, error) {

	user, err := s.repo.GetUserWithProfile(ctx, userID)
	if err != nil {
		return database.GetUserWithProfileRow{},
			fmt.Errorf("get user with profile: %w", err)
	}

	return user, nil
}

func (s *UserService) UpdateUser(
	ctx context.Context,
	params database.UpdateUserParams,
) (database.User, error) {
	if err := validators.ValidateUsername(params.Username); err != nil {
		return database.User{}, err
	}
	if err := validators.ValidateEmail(params.Email); err != nil {
		return database.User{}, err
	}

	currentUser, err := s.repo.GetUserByID(ctx, params.ID)
	if err != nil {
		return database.User{}, fmt.Errorf("get user: %w", err)
	}

	if params.Username != currentUser.Username {
		if err := s.guard.EnsureUsernameAvailable(ctx, params.Username); err != nil {
			return database.User{}, err
		}
	}

	if params.Email != currentUser.Email {
		if err := s.guard.EnsureEmailAvailable(ctx, params.Email); err != nil {
			return database.User{}, err
		}
	}

	user, err := s.repo.UpdateUser(ctx, params)
	if err != nil {
		return database.User{}, fmt.Errorf("update user: %w", err)
	}

	return user, nil
}

func (s *UserService) UpdateUserProfile(
	ctx context.Context,
	params database.UpdateUserProfileParams,
) (database.UserProfile, error) {
	if err := validators.ValidateFirstName(params.FirstName); err != nil {
		return database.UserProfile{}, err
	}
	if err := validators.ValidateLastName(params.LastName); err != nil {
		return database.UserProfile{}, err
	}
	if err := validators.ValidateUserBio(params.Bio.String); err != nil {
		return database.UserProfile{}, err
	}
	if err := validators.ValidateAvatarURL(params.AvatarUrl.String); err != nil {
		return database.UserProfile{}, err
	}

	_, err := s.repo.GetUserProfile(ctx, params.UserID)
	if err != nil {
		return database.UserProfile{},
			fmt.Errorf("get user profile: %w", err)
	}

	profile, err := s.repo.UpdateUserProfile(ctx, params)
	if err != nil {
		return database.UserProfile{},
			fmt.Errorf("update user profile: %w", err)
	}

	return profile, nil
}

func (s *UserService) UpdateUserAvatar(
	ctx context.Context,
	params database.UpdateUserAvatarParams,
) (database.UserProfile, error) {
	if err := validators.ValidateAvatarURL(params.AvatarUrl.String); err != nil {
		return database.UserProfile{}, err
	}

	_, err := s.guard.EnsureUserProfileExists(ctx, params.UserID)
	if err != nil {
		return database.UserProfile{}, err
	}

	profile, err := s.repo.UpdateUserAvatar(ctx, params)
	if err != nil {
		return database.UserProfile{},
			fmt.Errorf("update avatar: %w", err)
	}

	return profile, nil
}

func (s *UserService) UpdateUserPassword(
	ctx context.Context,
	params database.UpdateUserPasswordParams,
) error {
	if err := validators.ValidatePasswordHash(params.PasswordHash); err != nil {
		return err
	}

	_, err := s.guard.EnsureUserExists(ctx, params.ID)
	if err != nil {
		return err
	}

	_, err = s.repo.UpdateUserPassword(ctx, params)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}

func (s *UserService) DeleteUser(
	ctx context.Context,
	id pgtype.UUID,
) error {
	_, err := s.guard.EnsureUserExists(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}
