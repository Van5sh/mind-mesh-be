package helpers

import (
	"time"

	"example/hello/graph/model"
	"example/hello/internal/database"
)

func ProjectToModel(project database.Project) *model.Project {
	var description *string

	if project.Description.Valid {
		description = &project.Description.String
	}

	var archivedAt *time.Time

	if project.ArchivedAt.Valid {
		archivedAt = &project.ArchivedAt.Time
	}

	return &model.Project{
		ID:          project.ID.String(),
		Name:        project.Name,
		Description: description,
		Visibility:  model.ProjectVisibility(project.Visibility),
		Owner: &model.User{
			ID: project.OwnerID.String(),
		},
		ArchivedAt: archivedAt,
		CreatedAt:  project.CreatedAt.Time,
		UpdatedAt:  project.UpdatedAt.Time,
	}
}

// ProjectMemberRoleToDB maps the GraphQL-facing ProjectMemberRole enum
// (OWNER/ADMIN/MEMBER/VIEWER) onto the database's ProjectRole enum
// (OWNER/ADMIN/EDITOR/VIEWER). The two enums were defined independently
// and only differ in the "member" tier's name, so a plain string cast
// between them sends an invalid "MEMBER" value into the project_role
// Postgres enum.
func ProjectMemberRoleToDB(role model.ProjectMemberRole) database.ProjectRole {
	if role == model.ProjectMemberRoleMember {
		return database.ProjectRoleEDITOR
	}

	return database.ProjectRole(role)
}

func ProjectMemberToModel(member database.ProjectMember) *model.ProjectMember {
	return &model.ProjectMember{
		ID: member.ID.String(),
		User: &model.User{
			ID: member.UserID.String(),
		},
		Project: &model.Project{
			ID: member.ProjectID.String(),
		},
		Role:      model.ProjectRole(member.Role),
		CreatedAt: member.CreatedAt.Time,
		UpdatedAt: member.UpdatedAt.Time,
	}
}
