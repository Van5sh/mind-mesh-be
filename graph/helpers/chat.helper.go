package helpers

import (
	"example/hello/graph/model"
	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

func ChatToModel(chat database.Chat) *model.Chat {
	return &model.Chat{
		ID:             chat.ID.String(),
		Title:          NullableString(chat.Title),
		Type:           model.ChatType(chat.Type),
		Status:         model.ChatStatus(chat.Status),
		LastActivityAt: chat.LastActivityAt.Time,
		CreatedAt:      chat.CreatedAt.Time,
		UpdatedAt:      chat.UpdatedAt.Time,
		Project: &model.Project{
			ID: chat.ProjectID.String(),
		},
	}
}

func ChatMessageToModel(message database.ChatMessage) *model.ChatMessage {
	var sender *model.User

	if message.SenderID.Valid {
		sender = &model.User{
			ID: message.SenderID.String(),
		}
	}

	return &model.ChatMessage{
		ID:      message.ID.String(),
		Content: message.Content,
		Role:    model.MessageRole(message.Role),
		Sender:  sender,
		Chat: &model.Chat{
			ID: message.ChatID.String(),
		},
		CreatedAt: message.CreatedAt.Time,
		UpdatedAt: message.UpdatedAt.Time,
	}
}

func NullableString(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}
