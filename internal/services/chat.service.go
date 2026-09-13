package services

import (
	"context"
	"log"

	"example/hello/internal/apperrors"
	"example/hello/internal/database"
	"example/hello/internal/guards"
	"example/hello/internal/repository"
	"example/hello/internal/services/ai"
	"example/hello/internal/validators"

	"github.com/jackc/pgx/v5/pgtype"
)

type ChatService struct {
	repo   *repository.ChatRepository
	guard  *guards.ChatGuard
	aiChat *ai.Client
}

func NewChatService(
	repo *repository.ChatRepository,
	guard *guards.ChatGuard,
	aiChat *ai.Client,
) *ChatService {
	return &ChatService{
		repo:   repo,
		guard:  guard,
		aiChat: aiChat,
	}
}

// CreateChat creates a new chat.
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
		return database.Chat{}, apperrors.InternalError(
			"failed to create chat",
			err,
		)
	}

	return chat, nil
}

// GetChatByID gets a chat by ID.
func (s *ChatService) GetChatByID(
	ctx context.Context,
	id pgtype.UUID,
) (database.Chat, error) {
	if err := validators.ValidateUUID("chat id", id); err != nil {
		return database.Chat{}, err
	}

	return s.guard.EnsureChatExists(ctx, id)
}

// UpdateChat updates the chat title.
func (s *ChatService) UpdateChat(
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
		return database.Chat{}, apperrors.InternalError(
			"failed to update chat",
			err,
		)
	}

	return chat, nil
}

// UpdateChatStatus updates the status of a chat.
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
		return database.Chat{}, apperrors.InternalError(
			"failed to update chat status",
			err,
		)
	}

	return chat, nil
}

// UpdateChatType updates the type of a chat.
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
		return database.Chat{}, apperrors.InternalError(
			"failed to update chat type",
			err,
		)
	}

	return chat, nil
}

// DeleteChat deletes a chat.
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
		return apperrors.InternalError(
			"failed to delete chat",
			err,
		)
	}

	return nil
}

// CreateChatParticipant adds a user to a chat.
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

	if err := s.guard.EnsureUserNotInChat(
		ctx,
		params.ChatID,
		params.UserID,
	); err != nil {
		return database.ChatParticipant{}, err
	}

	participant, err := s.repo.CreateChatParticipant(ctx, params)
	if err != nil {
		return database.ChatParticipant{}, apperrors.InternalError(
			"failed to create chat participant",
			err,
		)
	}

	if err := s.repo.UpdateChatActivity(ctx, params.ChatID); err != nil {
		return database.ChatParticipant{}, apperrors.InternalError(
			"failed to update chat activity",
			err,
		)
	}

	return participant, nil
}

// GetChatParticipant gets one participant from a chat.
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

	return s.guard.EnsureChatParticipant(
		ctx,
		params.ChatID,
		params.UserID,
	)
}

// GetChatParticipants gets all participants in a chat.
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
		return nil, apperrors.InternalError(
			"failed to fetch chat participants",
			err,
		)
	}

	return participants, nil
}

// RemoveChatParticipant removes a participant from a chat.
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

	if _, err := s.guard.EnsureChatParticipant(
		ctx,
		params.ChatID,
		params.UserID,
	); err != nil {
		return err
	}

	if err := s.repo.RemoveChatParticipant(ctx, params); err != nil {
		return apperrors.InternalError(
			"failed to remove chat participant",
			err,
		)
	}

	if err := s.repo.UpdateChatActivity(ctx, params.ChatID); err != nil {
		return apperrors.InternalError(
			"failed to update chat activity",
			err,
		)
	}

	return nil
}

// CreateChatMessage creates a message in a chat.
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

	chat, err := s.guard.EnsureChatExists(ctx, params.ChatID)
	if err != nil {
		return database.ChatMessage{}, err
	}

	message, err := s.repo.CreateChatMessage(ctx, params)
	if err != nil {
		return database.ChatMessage{}, apperrors.InternalError(
			"failed to create chat message",
			err,
		)
	}

	if err := s.repo.UpdateChatActivity(ctx, params.ChatID); err != nil {
		return database.ChatMessage{}, apperrors.InternalError(
			"failed to update chat activity",
			err,
		)
	}

	// AI_ASSISTANT chats get an automatic reply to user messages. A
	// failure here must not fail message creation: the user's message
	// is already saved, and the AI companion being unavailable is a
	// degraded experience, not an error the caller should see.
	if chat.Type == database.ChatTypeAIASSISTANT && params.Role == database.MessageRoleUSER {
		s.generateAIReply(ctx, chat, params.Content)
	}

	return message, nil
}

// generateAIReply asks the AI service for an answer grounded in the
// chat's project documents and persists it as an AI-role message.
// Errors are logged, never returned - see CreateChatMessage.
func (s *ChatService) generateAIReply(
	ctx context.Context,
	chat database.Chat,
	question string,
) {
	if s.aiChat == nil {
		return
	}

	result, err := s.aiChat.AnswerChatQuestion(ctx, chat.ProjectID.String(), question)
	if err != nil {
		log.Printf("AI reply generation failed for chat %s: %v", chat.ID.String(), err)
		return
	}

	if _, err := s.repo.CreateChatMessage(ctx, database.CreateChatMessageParams{
		ChatID:  chat.ID,
		Role:    database.MessageRoleAI,
		Content: result.Answer,
	}); err != nil {
		log.Printf("failed to save AI reply for chat %s: %v", chat.ID.String(), err)
		return
	}

	if err := s.repo.UpdateChatActivity(ctx, chat.ID); err != nil {
		log.Printf("failed to update chat activity after AI reply for chat %s: %v", chat.ID.String(), err)
	}
}

// GetChatMessageByID gets a message by ID.
func (s *ChatService) GetChatMessageByID(
	ctx context.Context,
	id pgtype.UUID,
) (database.ChatMessage, error) {
	if err := validators.ValidateUUID("message id", id); err != nil {
		return database.ChatMessage{}, err
	}

	return s.guard.EnsureChatMessageExists(ctx, id)
}

// GetChatMessagesByChatID gets all messages in a chat.
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
		return nil, apperrors.InternalError(
			"failed to fetch chat messages",
			err,
		)
	}

	return messages, nil
}

// UpdateChatMessage updates a chat message.
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
		return database.ChatMessage{}, apperrors.InternalError(
			"failed to update chat message",
			err,
		)
	}

	if err := s.repo.UpdateChatActivity(ctx, existing.ChatID); err != nil {
		return database.ChatMessage{}, apperrors.InternalError(
			"failed to update chat activity",
			err,
		)
	}

	return message, nil
}

// DeleteChatMessage deletes a chat message.
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
		return apperrors.InternalError(
			"failed to delete chat message",
			err,
		)
	}

	if err := s.repo.UpdateChatActivity(ctx, message.ChatID); err != nil {
		return apperrors.InternalError(
			"failed to update chat activity",
			err,
		)
	}

	return nil
}

// GetChatsByProjectID gets all chats belonging to a project.
func (s *ChatService) GetChatsByProjectID(
	ctx context.Context,
	projectID pgtype.UUID,
) ([]database.Chat, error) {
	if err := validators.ValidateUUID("project id", projectID); err != nil {
		return nil, err
	}

	chats, err := s.repo.GetChatsByProjectID(ctx, projectID)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch project chats",
			err,
		)
	}

	return chats, nil
}

// GetChatsByUserID gets all chats for a user.
func (s *ChatService) GetChatsByUserID(
	ctx context.Context,
	userID pgtype.UUID,
) ([]database.Chat, error) {
	if err := validators.ValidateUUID("user id", userID); err != nil {
		return nil, err
	}

	chats, err := s.repo.GetChatsByUserID(ctx, userID)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch user chats",
			err,
		)
	}

	return chats, nil
}

// GetChatsByProjectAndUser gets chats for a user within a project.
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
		return nil, apperrors.InternalError(
			"failed to fetch chats by project and user",
			err,
		)
	}

	return chats, nil
}

// GetActiveChats gets active chats for a project.
func (s *ChatService) GetActiveChats(
	ctx context.Context,
	projectID pgtype.UUID,
) ([]database.Chat, error) {
	if err := validators.ValidateUUID("project id", projectID); err != nil {
		return nil, err
	}

	chats, err := s.repo.GetActiveChats(ctx, projectID)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch active chats",
			err,
		)
	}

	return chats, nil
}

// GetArchivedChats gets archived chats for a project.
func (s *ChatService) GetArchivedChats(
	ctx context.Context,
	projectID pgtype.UUID,
) ([]database.Chat, error) {
	if err := validators.ValidateUUID("project id", projectID); err != nil {
		return nil, err
	}

	chats, err := s.repo.GetArchivedChats(ctx, projectID)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch archived chats",
			err,
		)
	}

	return chats, nil
}

// GetLatestChatMessage gets the latest message in a chat.
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
		return database.ChatMessage{}, apperrors.InternalError(
			"failed to fetch latest chat message",
			err,
		)
	}

	return message, nil
}

// CheckUserInChat checks whether a user belongs to a chat.
func (s *ChatService) CheckUserInChat(
	ctx context.Context,
	params database.CheckUserInChatParams,
) (bool, error) {
	if err := validators.ValidateUUID("chat id", params.ChatID); err != nil {
		return false, err
	}

	if err := validators.ValidateUUID("user id", params.UserID); err != nil {
		return false, err
	}

	inChat, err := s.repo.CheckUserInChat(ctx, params)
	if err != nil {
		return false, apperrors.InternalError(
			"failed to check chat membership",
			err,
		)
	}

	return inChat, nil
}

// CreateMessageMention creates a mention for a message.
func (s *ChatService) CreateMessageMention(
	ctx context.Context,
	params database.CreateMessageMentionParams,
) (database.MessageMention, error) {
	if err := validators.ValidateUUID("message id", params.MessageID); err != nil {
		return database.MessageMention{}, err
	}

	if err := validators.ValidateUUID(
		"mentioned user id",
		params.MentionedUserID,
	); err != nil {
		return database.MessageMention{}, err
	}

	if _, err := s.guard.EnsureChatMessageExists(
		ctx,
		params.MessageID,
	); err != nil {
		return database.MessageMention{}, err
	}

	mention, err := s.repo.CreateMessageMention(ctx, params)
	if err != nil {
		return database.MessageMention{}, apperrors.InternalError(
			"failed to create message mention",
			err,
		)
	}

	return mention, nil
}

// GetMessageMentions gets users/messages mentioned by a message.
func (s *ChatService) GetMessageMentions(
	ctx context.Context,
	messageID pgtype.UUID,
) ([]database.GetMessageMentionsRow, error) {
	if err := validators.ValidateUUID("message id", messageID); err != nil {
		return nil, err
	}

	if _, err := s.guard.EnsureChatMessageExists(
		ctx,
		messageID,
	); err != nil {
		return nil, err
	}

	mentions, err := s.repo.GetMessageMentions(ctx, messageID)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch message mentions",
			err,
		)
	}

	return mentions, nil
}

// GetChatMessagesByIDs gets messages by a list of IDs.
func (s *ChatService) GetChatMessagesByIDs(
	ctx context.Context,
	ids []pgtype.UUID,
) ([]database.ChatMessage, error) {
	if err := validators.ValidateUUIDSlice(
		"message ids",
		ids,
		true,
	); err != nil {
		return nil, err
	}

	messages, err := s.repo.GetChatMessagesByIDs(ctx, ids)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch chat messages by ids",
			err,
		)
	}

	return messages, nil
}

// GetChatMessagesByRole gets messages by role.
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
		return nil, apperrors.InternalError(
			"failed to fetch chat messages by role",
			err,
		)
	}

	return messages, nil
}

// GetChatMessagesBySender gets messages from a particular sender.
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
		return nil, apperrors.InternalError(
			"failed to fetch chat messages by sender",
			err,
		)
	}

	return messages, nil
}

// GetChatMessagesWithSender gets messages including sender information.
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
		return nil, apperrors.InternalError(
			"failed to fetch chat messages with sender",
			err,
		)
	}

	return messages, nil
}

// GetChatParticipantsWithUsers gets participants with their user data.
func (s *ChatService) GetChatParticipantsWithUsers(
	ctx context.Context,
	chatID pgtype.UUID,
) ([]database.GetChatParticipantsWithUsersRow, error) {
	if err := validators.ValidateUUID("chat id", chatID); err != nil {
		return nil, err
	}

	if _, err := s.guard.EnsureChatExists(ctx, chatID); err != nil {
		return nil, err
	}

	participants, err := s.repo.GetChatParticipantsWithUsers(ctx, chatID)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to fetch chat participants with users",
			err,
		)
	}

	return participants, nil
}

// UpdateChatActivity updates the last activity timestamp.
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
		return apperrors.InternalError(
			"failed to update chat activity",
			err,
		)
	}

	return nil
}

// RenameChat renames a chat.
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
		return database.Chat{}, apperrors.InternalError(
			"failed to rename chat",
			err,
		)
	}

	return chat, nil
}

// SearchChatMessages searches messages.
func (s *ChatService) SearchChatMessages(
	ctx context.Context,
	params database.SearchChatMessagesParams,
) ([]database.ChatMessage, error) {
	if err := validators.ValidateUUID(
		"project id",
		params.ProjectID,
	); err != nil {
		return nil, err
	}

	if err := validators.ValidateOptionalMaxLength(
		"search text",
		params.Column2.String,
		500,
	); err != nil {
		return nil, err
	}

	messages, err := s.repo.SearchChatMessages(ctx, params)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to search chat messages",
			err,
		)
	}

	return messages, nil
}

// SearchChatsByTitle searches chats by title.
func (s *ChatService) SearchChatsByTitle(
	ctx context.Context,
	params database.SearchChatsByTitleParams,
) ([]database.Chat, error) {
	if err := validators.ValidateUUID(
		"project id",
		params.ProjectID,
	); err != nil {
		return nil, err
	}

	if err := validators.ValidateOptionalMaxLength(
		"chat title",
		params.Column2.String,
		100,
	); err != nil {
		return nil, err
	}

	chats, err := s.repo.SearchChatsByTitle(ctx, params)
	if err != nil {
		return nil, apperrors.InternalError(
			"failed to search chats by title",
			err,
		)
	}

	return chats, nil
}
