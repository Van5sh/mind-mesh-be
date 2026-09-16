package guards

import (
	"context"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

type ChatGuard struct {
	repo *repository.ChatRepository
}

func NewChatGuard(repo *repository.ChatRepository) *ChatGuard {
	return &ChatGuard{repo: repo}
}

func (g *ChatGuard) EnsureChatExists(ctx context.Context, chatID pgtype.UUID) (database.Chat, error) {
	chat, err := g.repo.GetChatByID(ctx, chatID)
	if err != nil {
		if isNoRows(err) {
			return database.Chat{}, apperrors.NotFoundError("chat not found")
		}
		return database.Chat{}, apperrors.InternalError("failed to fetch chat", err)
	}
	return chat, nil
}

func (g *ChatGuard) EnsureChatMessageExists(ctx context.Context, messageID pgtype.UUID) (database.ChatMessage, error) {
	message, err := g.repo.GetChatMessageByID(ctx, messageID)
	if err != nil {
		if isNoRows(err) {
			return database.ChatMessage{}, apperrors.NotFoundError("chat message not found")
		}
		return database.ChatMessage{}, apperrors.InternalError("failed to fetch chat message", err)
	}
	return message, nil
}

func (g *ChatGuard) EnsureChatBelongsToProject(ctx context.Context, chatID, projectID pgtype.UUID) (database.Chat, error) {
	chat, err := g.EnsureChatExists(ctx, chatID)
	if err != nil {
		return database.Chat{}, err
	}
	if !sameUUID(chat.ProjectID, projectID) {
		return database.Chat{}, apperrors.ForbiddenError("chat does not belong to project")
	}
	return chat, nil
}

func (g *ChatGuard) EnsureChatParticipant(ctx context.Context, chatID, userID pgtype.UUID) (database.ChatParticipant, error) {
	participant, err := g.repo.GetChatParticipant(ctx, database.GetChatParticipantParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		if isNoRows(err) {
			return database.ChatParticipant{}, apperrors.ForbiddenError("user is not a chat participant")
		}
		return database.ChatParticipant{}, apperrors.InternalError("failed to fetch chat participant", err)
	}
	return participant, nil
}

func (g *ChatGuard) EnsureUserNotInChat(ctx context.Context, chatID, userID pgtype.UUID) error {
	inChat, err := g.repo.CheckUserInChat(ctx, database.CheckUserInChatParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return apperrors.InternalError("failed to check chat participant", err)
	}
	if inChat {
		return apperrors.ConflictError("user is already in chat")
	}
	return nil
}
