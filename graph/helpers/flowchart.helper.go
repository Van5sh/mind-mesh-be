package helpers

import (
	"example/hello/graph/model"
	"example/hello/internal/database"
)

func FlowchartToModel(flowchart database.Flowchart) *model.Flowchart {
	var generatedBy *model.User

	if flowchart.GeneratedBy.Valid {
		generatedBy = &model.User{ID: flowchart.GeneratedBy.String()}
	}

	var sourceChat *model.Chat

	if flowchart.SourceChatID.Valid {
		sourceChat = &model.Chat{ID: flowchart.SourceChatID.String()}
	}

	return &model.Flowchart{
		ID:            flowchart.ID.String(),
		Name:          flowchart.Name,
		Data:          string(flowchart.Data),
		GeneratedBy:   generatedBy,
		GeneratedByAi: flowchart.GeneratedByAi.Bool,
		Status:        model.FlowchartStatus(flowchart.Status),
		SourceChat:    sourceChat,
		CreatedAt:     flowchart.CreatedAt.Time,
		UpdatedAt:     flowchart.UpdatedAt.Time,
		Project: &model.Project{
			ID: flowchart.ProjectID.String(),
		},
	}
}
