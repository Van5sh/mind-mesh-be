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

type ChatService struct {
	repo  *repository.ChatRepository
	guard *guards.ChatGuard
}

func NewChatService(repo *repository.ChatRepository, guard *guards.ChatGuard) *ChatService {
	return &ChatService{
		repo:  repo,
		guard: guard,
	}
}

func (s *ChatService) CreateChat(
	ctx context.Context,
	params database.CreateChatParams,
) (database.Chat, error) {
	if err := validators.ValidateUUID("project id", params.ProjectID); err != nil {
		return database.Chat{}, err
	}
	if err := validators.ValidateChatTitle(params.Title.String); err != nil {
		return database.Chat{}, err
	}

	chat, err := s.repo.CreateChat(ctx, params)
	if err != nil {
		return database.Chat{}, apperrors.InternalError("failed to create chat", err)
	}

	return chat, nil
}

func (s *ChatService) GetChatByID(
	ctx context.Context,
	id pgtype.UUID,
) (database.Chat, error) {
	if err := validators.ValidateUUID("chat id", id); err != nil {
		return database.Chat{}, err
	}

	return s.guard.EnsureChatExists(ctx, id)
}

func (s *ChatService) DeleteChat(
	ctx context.Context,
	id pgtype.UUID,
) error {
	if err := validators.ValidateUUID("chat id", id); err != nil {
		return err
	}
	if _, err := s.guard.EnsureChatExists(ctx, id); err != nil {
		return err
	}

	if err := s.repo.DeleteChat(ctx, id); err != nil {
		return apperrors.InternalError("failed to delete chat", err)
	}

	return nil
}

func (s *ChatService) CreateChatMessage(
	ctx context.Context,
	params database.CreateChatMessageParams,
) (database.ChatMessage, error) {
	if err := validators.ValidateUUID("chat id", params.ChatID); err != nil {
		return database.ChatMessage{}, err
	}
	if params.SenderID.Valid {
		if err := validators.ValidateUUID("sender id", params.SenderID); err != nil {
			return database.ChatMessage{}, err
		}
	}
	if err := validators.ValidateChatMessageContent(params.Content); err != nil {
		return database.ChatMessage{}, err
	}
	if _, err := s.guard.EnsureChatExists(ctx, params.ChatID); err != nil {
		return database.ChatMessage{}, err
	}

	message, err := s.repo.CreateChatMessage(ctx, params)
	if err != nil {
		return database.ChatMessage{}, apperrors.InternalError("failed to create chat message", err)
	}

	if err := s.repo.UpdateChatActivity(ctx, params.ChatID); err != nil {
		return database.ChatMessage{}, apperrors.InternalError("failed to update chat activity", err)
	}

	return message, nil
}

func (s *ChatService) GetChatMessageByID(
	ctx context.Context,
	id pgtype.UUID,
) (database.ChatMessage, error) {
	if err := validators.ValidateUUID("message id", id); err != nil {
		return database.ChatMessage{}, err
	}

	return s.guard.EnsureChatMessageExists(ctx, id)
}

func (s *ChatService) DeleteChatMessage(
	ctx context.Context,
	id pgtype.UUID,
) error {
	if err := validators.ValidateUUID("message id", id); err != nil {
		return err
	}
	message, err := s.guard.EnsureChatMessageExists(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteChatMessage(ctx, id); err != nil {
		return apperrors.InternalError("failed to delete chat message", err)
	}
	if err := s.repo.UpdateChatActivity(ctx, message.ChatID); err != nil {
		return apperrors.InternalError("failed to update chat activity", err)
	}

	return nil
}

func (s *ChatService) GetChatMessagesByChatID(
	ctx context.Context,
	chatID pgtype.UUID,
) ([]database.ChatMessage, error) {
	if err := validators.ValidateUUID("chat id", chatID); err != nil {
		return nil, err
	}
	if _, err := s.guard.EnsureChatExists(ctx, chatID); err != nil {
		return nil, err
	}

	messages, err := s.repo.GetChatMessagesByChatID(ctx, chatID)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch chat messages", err)
	}

	return messages, nil
}

func (s *ChatService) GetChatParticipants(
	ctx context.Context,
	chatID pgtype.UUID,
) ([]database.ChatParticipant, error) {
	if err := validators.ValidateUUID("chat id", chatID); err != nil {
		return nil, err
	}
	if _, err := s.guard.EnsureChatExists(ctx, chatID); err != nil {
		return nil, err
	}

	participants, err := s.repo.GetChatParticipants(ctx, chatID)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch chat participants", err)
	}

	return participants, nil
}

func (s *ChatService) CreateChatParticipant(
	ctx context.Context,
	params database.CreateChatParticipantParams,
) (database.ChatParticipant, error) {
	if err := validators.ValidateUUID("chat id", params.ChatID); err != nil {
		return database.ChatParticipant{}, err
	}
	if err := validators.ValidateUUID("user id", params.UserID); err != nil {
		return database.ChatParticipant{}, err
	}
	if _, err := s.guard.EnsureChatExists(ctx, params.ChatID); err != nil {
		return database.ChatParticipant{}, err
	}
	if err := s.guard.EnsureUserNotInChat(ctx, params.ChatID, params.UserID); err != nil {
		return database.ChatParticipant{}, err
	}

	participant, err := s.repo.CreateChatParticipant(ctx, params)
	if err != nil {
		return database.ChatParticipant{}, apperrors.InternalError("failed to create chat participant", err)
	}
	if err := s.repo.UpdateChatActivity(ctx, params.ChatID); err != nil {
		return database.ChatParticipant{}, apperrors.InternalError("failed to update chat activity", err)
	}

	return participant, nil
}

func (s *ChatService) GetChatsByProjectID(
	ctx context.Context,
	projectID pgtype.UUID,
) ([]database.Chat, error) {
	if err := validators.ValidateUUID("project id", projectID); err != nil {
		return nil, err
	}

	chats, err := s.repo.GetChatsByProjectID(ctx, projectID)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch project chats", err)
	}

	return chats, nil
}

func (s *ChatService) GetChatsByUserID(
	ctx context.Context,
	userID pgtype.UUID,
) ([]database.Chat, error) {
	if err := validators.ValidateUUID("user id", userID); err != nil {
		return nil, err
	}

	chats, err := s.repo.GetChatsByUserID(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch user chats", err)
	}

	return chats, nil
}

func (s *ChatService) GetChatsByProjectAndUser(
	ctx context.Context,
	params database.GetChatsByProjectAndUserParams,
) ([]database.Chat, error) {
	if err := validators.ValidateUUID("project id", params.ProjectID); err != nil {
		return nil, err
	}
	if err := validators.ValidateUUID("user id", params.UserID); err != nil {
		return nil, err
	}

	chats, err := s.repo.GetChatsByProjectAndUser(ctx, params)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch chats by project and user", err)
	}

	return chats, nil
}

func (s *ChatService) GetActiveChats(
	ctx context.Context,
	projectID pgtype.UUID,
) ([]database.Chat, error) {
	if err := validators.ValidateUUID("project id", projectID); err != nil {
		return nil, err
	}

	chats, err := s.repo.GetActiveChats(ctx, projectID)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch active chats", err)
	}

	return chats, nil
}

func (s *ChatService) GetArchivedChats(
	ctx context.Context,
	projectID pgtype.UUID,
) ([]database.Chat, error) {
	if err := validators.ValidateUUID("project id", projectID); err != nil {
		return nil, err
	}

	chats, err := s.repo.GetArchivedChats(ctx, projectID)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch archived chats", err)
	}

	return chats, nil
}

func (s *ChatService) UpdateChatStatus(
	ctx context.Context,
	params database.UpdateChatStatusParams,
) (database.Chat, error) {
	if err := validators.ValidateUUID("chat id", params.ID); err != nil {
		return database.Chat{}, err
	}
	if _, err := s.guard.EnsureChatExists(ctx, params.ID); err != nil {
		return database.Chat{}, err
	}

	chat, err := s.repo.UpdateChatStatus(ctx, params)
	if err != nil {
		return database.Chat{}, apperrors.InternalError("failed to update chat status", err)
	}

	return chat, nil
}

func (s *ChatService) UpdateChatType(
	ctx context.Context,
	params database.UpdateChatTypeParams,
) (database.Chat, error) {
	if err := validators.ValidateUUID("chat id", params.ID); err != nil {
		return database.Chat{}, err
	}
	if _, err := s.guard.EnsureChatExists(ctx, params.ID); err != nil {
		return database.Chat{}, err
	}

	chat, err := s.repo.UpdateChatType(ctx, params)
	if err != nil {
		return database.Chat{}, apperrors.InternalError("failed to update chat type", err)
	}

	return chat, nil
}

func (s *ChatService) GetLatestChatMessage(
	ctx context.Context,
	chatID pgtype.UUID,
) (database.ChatMessage, error) {
	if err := validators.ValidateUUID("chat id", chatID); err != nil {
		return database.ChatMessage{}, err
	}
	if _, err := s.guard.EnsureChatExists(ctx, chatID); err != nil {
		return database.ChatMessage{}, err
	}

	message, err := s.repo.GetLatestChatMessage(ctx, chatID)
	if err != nil {
		return database.ChatMessage{}, apperrors.InternalError("failed to fetch latest chat message", err)
	}

	return message, nil
}

func (s *ChatService) CreateMessageMention(
	ctx context.Context,
	params database.CreateMessageMentionParams,
) (database.MessageMention, error) {
	if err := validators.ValidateUUID("message id", params.MessageID); err != nil {
		return database.MessageMention{}, err
	}
	if err := validators.ValidateUUID("mentioned user id", params.MentionedUserID); err != nil {
		return database.MessageMention{}, err
	}
	if _, err := s.guard.EnsureChatMessageExists(ctx, params.MessageID); err != nil {
		return database.MessageMention{}, err
	}

	mention, err := s.repo.CreateMessageMention(ctx, params)
	if err != nil {
		return database.MessageMention{}, apperrors.InternalError("failed to create message mention", err)
	}

	return mention, nil
}

func (s *ChatService) GetMessageMentions(
	ctx context.Context,
	messageID pgtype.UUID,
) ([]database.GetMessageMentionsRow, error) {
	if err := validators.ValidateUUID("message id", messageID); err != nil {
		return nil, err
	}
	if _, err := s.guard.EnsureChatMessageExists(ctx, messageID); err != nil {
		return nil, err
	}

	mentions, err := s.repo.GetMessageMentions(ctx, messageID)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch message mentions", err)
	}

	return mentions, nil
}

func (s *ChatService) GetChatMessagesByIDs(
	ctx context.Context,
	ids []pgtype.UUID,
) ([]database.ChatMessage, error) {
	if err := validators.ValidateUUIDSlice("message ids", ids, true); err != nil {
		return nil, err
	}

	messages, err := s.repo.GetChatMessagesByIDs(ctx, ids)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch chat messages by ids", err)
	}

	return messages, nil
}

func (s *ChatService) GetChatMessagesByRole(
	ctx context.Context,
	params database.GetChatMessagesByRoleParams,
) ([]database.ChatMessage, error) {
	if err := validators.ValidateUUID("chat id", params.ChatID); err != nil {
		return nil, err
	}
	if _, err := s.guard.EnsureChatExists(ctx, params.ChatID); err != nil {
		return nil, err
	}

	messages, err := s.repo.GetChatMessagesByRole(ctx, params)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch chat messages by role", err)
	}

	return messages, nil
}

func (s *ChatService) GetChatMessagesBySender(
	ctx context.Context,
	params database.GetChatMessagesBySenderParams,
) ([]database.ChatMessage, error) {
	if err := validators.ValidateUUID("chat id", params.ChatID); err != nil {
		return nil, err
	}
	if err := validators.ValidateUUID("sender id", params.SenderID); err != nil {
		return nil, err
	}
	if _, err := s.guard.EnsureChatExists(ctx, params.ChatID); err != nil {
		return nil, err
	}

	messages, err := s.repo.GetChatMessagesBySender(ctx, params)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch chat messages by sender", err)
	}

	return messages, nil
}

func (s *ChatService) GetChatMessagesWithSender(
	ctx context.Context,
	chatID pgtype.UUID,
) ([]database.GetChatMessagesWithSenderRow, error) {
	if err := validators.ValidateUUID("chat id", chatID); err != nil {
		return nil, err
	}
	if _, err := s.guard.EnsureChatExists(ctx, chatID); err != nil {
		return nil, err
	}

	messages, err := s.repo.GetChatMessagesWithSender(ctx, chatID)
	if err != nil {
		return nil, apperrors.InternalError("failed to fetch chat messages with sender", err)
	}

	return messages, nil
}

func (s *ChatService) GetChatParticipant(
	ctx context.Context,
	params database.GetChatParticipantParams,
) (database.ChatParticipant, error) {
	if err := validators.ValidateUUID("chat id", params.ChatID); err != nil {
		return database.ChatParticipant{}, err
	}
	if err := validators.ValidateUUID("user id", params.UserID); err != nil {
		return database.ChatParticipant{}, err
	}

	return s.guard.EnsureChatParticipant(ctx, params.ChatID, params.UserID)
}

func (s *ChatService) RemoveChatParticipant(
	ctx context.Context,
	params database.RemoveChatParticipantParams,
) error {
	if err := validators.ValidateUUID("chat id", params.ChatID); err != nil {
		return err
	}
	if err := validators.ValidateUUID("user id", params.UserID); err != nil {
		return err
	}
	if _, err := s.guard.EnsureChatParticipant(ctx, params.ChatID, params.UserID); err != nil {
		return err
	}

	if err := s.repo.RemoveChatParticipant(ctx, params); err != nil {
		return apperrors.InternalError("failed to remove chat participant", err)
	}
	if err := s.repo.UpdateChatActivity(ctx, params.ChatID); err != nil {
		return apperrors.InternalError("failed to update chat activity", err)
	}

	return nil
}

func (s *ChatService) RenameChat(
	ctx context.Context,
	params database.RenameChatParams,
) (database.Chat, error) {
	if err := validators.ValidateUUID("chat id", params.ID); err != nil {
		return database.Chat{}, err
	}
	if err := validators.ValidateChatTitle(params.Title.String); err != nil {
		return database.Chat{}, err
	}
	if _, err := s.guard.EnsureChatExists(ctx, params.ID); err != nil {
		return database.Chat{}, err
	}

	chat, err := s.repo.RenameChat(ctx, params)
	if err != nil {
		return database.Chat{}, apperrors.InternalError("failed to rename chat", err)
	}

	return chat, nil
}

func (s *ChatService) SearchChatMessages(
	ctx context.Context,
	params database.SearchChatMessagesParams,
) ([]database.ChatMessage, error) {
	if err := validators.ValidateUUID("project id", params.ProjectID); err != nil {
		return nil, err
	}
	if err := validators.ValidateOptionalMaxLength("search text", params.Column2.String, 500); err != nil {
		return nil, err
	}

	messages, err := s.repo.SearchChatMessages(ctx, params)
	if err != nil {
		return nil, apperrors.InternalError("failed to search chat messages", err)
	}

	return messages, nil
}

func (s *ChatService) SearchChatsByTitle(
	ctx context.Context,
	params database.SearchChatsByTitleParams,
) ([]database.Chat, error) {
	if err := validators.ValidateUUID("project id", params.ProjectID); err != nil {
		return nil, err
	}
	if err := validators.ValidateOptionalMaxLength("chat title", params.Column2.String, 100); err != nil {
		return nil, err
	}

	chats, err := s.repo.SearchChatsByTitle(ctx, params)
	if err != nil {
		return nil, apperrors.InternalError("failed to search chats by title", err)
	}

	return chats, nil
}

func (s *ChatService) UpdateChatActivity(
	ctx context.Context,
	id pgtype.UUID,
) error {
	if err := validators.ValidateUUID("chat id", id); err != nil {
		return err
	}
	if _, err := s.guard.EnsureChatExists(ctx, id); err != nil {
		return err
	}

	if err := s.repo.UpdateChatActivity(ctx, id); err != nil {
		return apperrors.InternalError("failed to update chat activity", err)
	}

	return nil
}

func (s *ChatService) UpdateChatMessage(
	ctx context.Context,
	params database.UpdateChatMessageParams,
) (database.ChatMessage, error) {
	if err := validators.ValidateUUID("message id", params.ID); err != nil {
		return database.ChatMessage{}, err
	}
	if err := validators.ValidateChatMessageContent(params.Content); err != nil {
		return database.ChatMessage{}, err
	}
	existing, err := s.guard.EnsureChatMessageExists(ctx, params.ID)
	if err != nil {
		return database.ChatMessage{}, err
	}

	message, err := s.repo.UpdateChatMessage(ctx, params)
	if err != nil {
		return database.ChatMessage{}, apperrors.InternalError("failed to update chat message", err)
	}
	if err := s.repo.UpdateChatActivity(ctx, existing.ChatID); err != nil {
		return database.ChatMessage{}, apperrors.InternalError("failed to update chat activity", err)
	}

	return message, nil
}
