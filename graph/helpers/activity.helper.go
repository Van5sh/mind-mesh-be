package helpers

import (
	"example/hello/graph/model"
	"example/hello/internal/database"
)

func ActivityLogToModel(activity database.ActivityLog) *model.ActivityLog {
	var project *model.Project

	if activity.ProjectID.Valid {
		project = &model.Project{ID: activity.ProjectID.String()}
	}

	var user *model.User

	if activity.UserID.Valid {
		user = &model.User{ID: activity.UserID.String()}
	}

	var entityID *string

	if activity.EntityID.Valid {
		id := activity.EntityID.String()
		entityID = &id
	}

	return &model.ActivityLog{
		ID:         activity.ID.String(),
		Project:    project,
		User:       user,
		Action:     activity.Action,
		EntityType: NullableString(activity.EntityType),
		EntityID:   entityID,
		CreatedAt:  activity.CreatedAt.Time,
	}
}
