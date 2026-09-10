package graph

import (
	"context"
	"fmt"

	"example/hello/graph/model"
	"example/hello/internal/apperrors"
	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
	user, err := r.App.Services.User.CreateUser(ctx, database.CreateUserParams{ID: newUUID(), Username: input.Username, Email: input.Email, PasswordHash: input.PasswordHash})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	profile, err := r.App.Services.User.GetUserProfile(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get created user profile: %w", err)
	}
	return userModel(user, &profile), nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*model.User, error) {
	id, err := parseUUID(input.UserID)
	if err != nil {
		return nil, err
	}
	user, err := r.App.Services.User.UpdateUser(ctx, database.UpdateUserParams{ID: id, Username: input.Username, Email: input.Email})
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	profile, err := r.App.Services.User.GetUserProfile(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}
	return userModel(user, &profile), nil
}

func (r *mutationResolver) UpdateUserPassword(ctx context.Context, input model.UpdateUserPasswordInput) (bool, error) {
	id, err := parseUUID(input.UserID)
	if err != nil {
		return false, err
	}
	err = r.App.Services.User.UpdateUserPassword(ctx, database.UpdateUserPasswordParams{ID: id, PasswordHash: input.PasswordHash})
	if err != nil {
		return false, fmt.Errorf("update user password: %w", err)
	}
	return true, nil
}

func (r *mutationResolver) DeleteUser(ctx context.Context, userID string) (bool, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return false, err
	}
	if err := r.App.Services.User.DeleteUser(ctx, id); err != nil {
		return false, fmt.Errorf("delete user: %w", err)
	}
	return true, nil
}

func (r *mutationResolver) CreateUserProfile(ctx context.Context, input model.CreateUserProfileInput) (*model.UserProfile, error) {
	id, err := parseUUID(input.UserID)
	if err != nil {
		return nil, err
	}
	profile, err := r.App.Services.User.CreateUserProfile(ctx, database.CreateUserProfileParams{UserID: id, FirstName: input.FirstName, LastName: input.LastName, Bio: textValue(input.Bio), AvatarUrl: textValue(input.AvatarURL)})
	if err != nil {
		return nil, fmt.Errorf("create user profile: %w", err)
	}
	return r.userProfileModel(ctx, profile)
}

func (r *mutationResolver) UpdateUserProfile(ctx context.Context, input model.UpdateUserProfileInput) (*model.UserProfile, error) {
	id, err := parseUUID(input.UserID)
	if err != nil {
		return nil, err
	}
	profile, err := r.App.Services.User.UpdateUserProfile(ctx, database.UpdateUserProfileParams{UserID: id, FirstName: input.FirstName, LastName: input.LastName, Bio: textValue(input.Bio), AvatarUrl: textValue(input.AvatarURL)})
	if err != nil {
		return nil, fmt.Errorf("update user profile: %w", err)
	}
	return r.userProfileModel(ctx, profile)
}

func (r *mutationResolver) UpdateUserAvatar(ctx context.Context, userID string, avatarURL *string) (*model.UserProfile, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}
	profile, err := r.App.Services.User.UpdateUserAvatar(ctx, database.UpdateUserAvatarParams{UserID: id, AvatarUrl: textValue(avatarURL)})
	if err != nil {
		return nil, fmt.Errorf("update user avatar: %w", err)
	}
	return r.userProfileModel(ctx, profile)
}

func (r *queryResolver) Me(context.Context) (*model.User, error) {
	return nil, apperrors.UnauthorizedError("authentication is not configured")
}

func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
	userID, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	user, err := r.App.Services.User.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	profile, err := r.App.Services.User.GetUserProfile(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}
	return userModel(user, &profile), nil
}

func (r *queryResolver) UserProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}
	profile, err := r.App.Services.User.GetUserProfile(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}
	return r.userProfileModel(ctx, profile)
}

func (r *queryResolver) AllUsers(ctx context.Context) ([]*model.User, error) {
	users, err := r.App.Services.User.GetAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}
	return r.userModels(ctx, users)
}

func (r *queryResolver) Users(ctx context.Context, ids []string) ([]*model.User, error) {
	userIDs := make([]pgtype.UUID, 0, len(ids))
	for _, value := range ids {
		id, err := parseUUID(value)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}
	users, err := r.App.Services.User.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}
	return r.userModels(ctx, users)
}

func (r *Resolver) userModels(ctx context.Context, users []database.User) ([]*model.User, error) {
	result := make([]*model.User, 0, len(users))
	for _, user := range users {
		profile, err := r.App.Services.User.GetUserProfile(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("get user profile: %w", err)
		}
		result = append(result, userModel(user, &profile))
	}
	return result, nil
}

func (r *Resolver) userProfileModel(ctx context.Context, profile database.UserProfile) (*model.UserProfile, error) {
	user, err := r.App.Services.User.GetUserByID(ctx, profile.UserID)
	if err != nil {
		return nil, fmt.Errorf("get profile user: %w", err)
	}
	return profileModel(profile, userModel(user, nil)), nil
}

func userModel(user database.User, profile *database.UserProfile) *model.User {
	result := userModelWithoutProfile(user)
	if profile != nil {
		result.Profile = profileModel(*profile, userModelWithoutProfile(user))
	}
	return result
}

func userModelWithoutProfile(user database.User) *model.User {
	return &model.User{ID: user.ID.String(), Username: user.Username, Email: user.Email, OwnedProjects: []*model.Project{}, ProjectMemberships: []*model.ProjectMember{}, UploadedFiles: []*model.File{}, SharedFiles: []*model.FileShare{}, CreatedAt: user.CreatedAt.Time, UpdatedAt: user.UpdatedAt.Time}
}

func profileModel(profile database.UserProfile, user *model.User) *model.UserProfile {
	return &model.UserProfile{User: user, FirstName: profile.FirstName, LastName: profile.LastName, Bio: textPointer(profile.Bio), AvatarURL: textPointer(profile.AvatarUrl), CreatedAt: profile.CreatedAt.Time, UpdatedAt: profile.UpdatedAt.Time}
}

func textValue(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
