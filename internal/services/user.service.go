package services

import (
	"context"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/guards"
	"example/hello/internal/repository"
	"example/hello/internal/validators"

	"github.com/jackc/pgx/v5/pgtype"
)

type UserService struct {
	repo  *repository.UserRepository
	guard *guards.UserGuard
}

func NewUserService(repo *repository.UserRepository, guard *guards.UserGuard) *UserService {
	return &UserService{
		repo:  repo,
		guard: guard,
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

	if err := s.guard.EnsureUsernameAvailable(ctx, params.Username); err != nil {
		return database.User{}, err
	}

	if err := s.guard.EnsureEmailAvailable(ctx, params.Email); err != nil {
		return database.User{}, err
	}

	user, err := s.repo.CreateUser(ctx, params)
	if err != nil {
		return database.User{}, apperrors.InternalError(
			"failed to create user",
			err,
		)
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

	if _, err := s.guard.EnsureUserExists(ctx, params.UserID); err != nil {
		return database.UserProfile{}, err
	}

	if _, err := s.guard.EnsureUserProfileExists(ctx, params.UserID); err == nil {
		profile, updateErr := s.repo.UpdateUserProfile(ctx, database.UpdateUserProfileParams(params))
		if updateErr != nil {
			return database.UserProfile{}, apperrors.InternalError("failed to update user profile", updateErr)
		}
		return profile, nil
	} else if !apperrors.IsCode(err, apperrors.NotFound) {
		return database.UserProfile{}, err
	}

	profile, err := s.repo.CreateUserProfile(ctx, params)
	if err != nil {
		return database.UserProfile{}, apperrors.InternalError("failed to create user profile", err)
	}

	return profile, nil
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]database.User, error) {
	users, err := s.repo.GetAllUsers(ctx)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch users", err)
	}
	return users, nil
}

func (s *UserService) GetUserByID(
	ctx context.Context,
	id pgtype.UUID,
) (database.User, error) {
	return s.guard.EnsureUserExists(ctx, id)
}

func (s *UserService) GetUserByEmail(
	ctx context.Context,
	email string,
) (database.User, error) {
	if err := validators.ValidateEmail(email); err != nil {
		return database.User{}, err
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return database.User{}, apperrors.InternalError("failed to fetch user by email", err)
	}

	return user, nil
}

func (s *UserService) GetUserByUsername(
	ctx context.Context,
	username string,
) (database.User, error) {
	if err := validators.ValidateUsername(username); err != nil {
		return database.User{}, err
	}

	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return database.User{}, apperrors.InternalError("failed to fetch user by username", err)
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
	if err := validators.ValidateUUIDSlice("user ids", ids, true); err != nil {
		return nil, err
	}

	users, err := s.repo.GetUsersByIDs(ctx, ids)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch users", err)
	}

	return users, nil
}

func (s *UserService) GetUserProfile(
	ctx context.Context,
	userID pgtype.UUID,
) (database.UserProfile, error) {
	return s.guard.EnsureUserProfileExists(ctx, userID)
}

func (s *UserService) GetUserWithProfile(
	ctx context.Context,
	userID pgtype.UUID,
) (database.GetUserWithProfileRow, error) {
	if _, err := s.guard.EnsureUserExists(ctx, userID); err != nil {
		return database.GetUserWithProfileRow{}, err
	}

	user, err := s.repo.GetUserWithProfile(ctx, userID)
	if err != nil {
		return database.GetUserWithProfileRow{}, apperrors.InternalError("failed to fetch user with profile", err)
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

	currentUser, err := s.guard.EnsureUserExists(ctx, params.ID)
	if err != nil {
		return database.User{}, err
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
		return database.User{}, apperrors.InternalError("failed to update user", err)
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

	if _, err := s.guard.EnsureUserProfileExists(ctx, params.UserID); err != nil {
		return database.UserProfile{}, err
	}

	profile, err := s.repo.UpdateUserProfile(ctx, params)
	if err != nil {
		return database.UserProfile{}, apperrors.InternalError("failed to update user profile", err)
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

	if _, err := s.guard.EnsureUserProfileExists(ctx, params.UserID); err != nil {
		return database.UserProfile{}, err
	}

	profile, err := s.repo.UpdateUserAvatar(ctx, params)
	if err != nil {
		return database.UserProfile{}, apperrors.InternalError("failed to update user avatar", err)
	}

	return profile, nil
}

func (s *UserService) DeleteUser(
	ctx context.Context,
	id pgtype.UUID,
) error {
	if _, err := s.guard.EnsureUserExists(ctx, id); err != nil {
		return err
	}

	if err := s.repo.DeleteUser(ctx, id); err != nil {
		return apperrors.InternalError("failed to delete user", err)
	}

	return nil
}

// GetUserProfilesByUserIDs fetches the profiles of many users in one query
// (users with no profile are absent). It exists for the User.profile
// dataloader.
func (s *UserService) GetUserProfilesByUserIDs(
	ctx context.Context,
	userIDs []pgtype.UUID,
) ([]database.UserProfile, error) {
	return fetchRows(ctx, userIDs, "user profiles", s.repo.GetUserProfilesByUserIDs)
}
