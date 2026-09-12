package helpers

import (
	"example/hello/graph/model"
	"example/hello/internal/database"
)

func UserToModel(user database.User) *model.User {
	return &model.User{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}
}

func UserProfileToModel(profile database.UserProfile) *model.UserProfile {
	return &model.UserProfile{
		User: &model.User{
			ID: profile.UserID.String(),
		},
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		Bio:       NullableString(profile.Bio),
		AvatarURL: NullableString(profile.AvatarUrl),
		CreatedAt: profile.CreatedAt.Time,
		UpdatedAt: profile.UpdatedAt.Time,
	}
}
