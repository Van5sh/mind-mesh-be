package helpers

import (
	"example/hello/graph/model"
	"example/hello/internal/database"
)

func ProjectToModel(project database.Project) *model.Project {
	var description *string

	if project.Description.Valid {
		description = &project.Description.String
	}

	return &model.Project{
		ID:          project.ID.String(),
		Name:        project.Name,
		Description: description,
	}
}

func ProjectMemberToModel(member database.ProjectMember) *model.ProjectMember {
	return &model.ProjectMember{
		ID: member.ProjectID.String(),
		User: &model.User{
			ID: member.ID.String(),
		},
		Role: model.ProjectRole(member.Role),
	}
}
