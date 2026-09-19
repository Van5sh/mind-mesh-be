package repository

import (
	"context"
	"example/hello/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

type ChatRepository struct {
	q *database.Queries
}

func NewChatRepository(q *database.Queries) *ChatRepository {
	return &ChatRepository{
		q: q,
	}
}

func (r *ChatRepository) CheckUserInChat(ctx context.Context, params database.CheckUserInChatParams) (bool, error) {
	return r.q.CheckUserInChat(ctx, params)
}

func (r *ChatRepository) CreateChat(ctx context.Context, params database.CreateChatParams) (database.Chat, error) {
	return r.q.CreateChat(ctx, params)
}

func (r *ChatRepository) CreateChatAIMetadata(ctx context.Context, params database.CreateChatAIMetadataParams) (database.ChatAiMetadatum, error) {
	return r.q.CreateChatAIMetadata(ctx, params)
}

func (r *ChatRepository) CreateChatMessage(ctx context.Context, params database.CreateChatMessageParams) (database.ChatMessage, error) {
	return r.q.CreateChatMessage(ctx, params)
}

func (r *ChatRepository) CreateChatParticipant(ctx context.Context, params database.CreateChatParticipantParams) (database.ChatParticipant, error) {
	return r.q.CreateChatParticipant(ctx, params)
}

func (r *ChatRepository) CreateMessageMention(ctx context.Context, params database.CreateMessageMentionParams) (database.MessageMention, error) {
	return r.q.CreateMessageMention(ctx, params)
}

func (r *ChatRepository) DeleteChat(ctx context.Context, id pgtype.UUID) error {
	return r.q.DeleteChat(ctx, id)
}

func (r *ChatRepository) DeleteChatAIMetadata(ctx context.Context, messageID pgtype.UUID) error {
	return r.q.DeleteChatAIMetadata(ctx, messageID)
}

func (r *ChatRepository) DeleteChatMessage(ctx context.Context, id pgtype.UUID) error {
	return r.q.DeleteChatMessage(ctx, id)
}

func (r *ChatRepository) DeleteMessageMentions(ctx context.Context, messageID pgtype.UUID) error {
	return r.q.DeleteMessageMentions(ctx, messageID)
}

func (r *ChatRepository) GeneratingChats(ctx context.Context, projectID pgtype.UUID) ([]database.Chat, error) {
	return r.q.GeneratingChats(ctx, projectID)
}

func (r *ChatRepository) GetActiveChats(ctx context.Context, projectID pgtype.UUID) ([]database.Chat, error) {
	return r.q.GetActiveChats(ctx, projectID)
}

func (r *ChatRepository) GetArchivedChatMessages(ctx context.Context, chatID pgtype.UUID) ([]database.ChatMessage, error) {
	return r.q.GetArchivedChatMessages(ctx, chatID)
}

func (r *ChatRepository) GetArchivedChats(ctx context.Context, projectID pgtype.UUID) ([]database.Chat, error) {
	return r.q.GetArchivedChats(ctx, projectID)
}

func (r *ChatRepository) GetChatAIMetadata(ctx context.Context, messageID pgtype.UUID) (database.ChatAiMetadatum, error) {
	return r.q.GetChatAIMetadata(ctx, messageID)
}

func (r *ChatRepository) GetChatByID(ctx context.Context, id pgtype.UUID) (database.Chat, error) {
	return r.q.GetChatByID(ctx, id)
}

func (r *ChatRepository) GetChatMessageByID(ctx context.Context, id pgtype.UUID) (database.ChatMessage, error) {
	return r.q.GetChatMessageByID(ctx, id)
}

func (r *ChatRepository) GetChatMessagesByChatID(ctx context.Context, chatID pgtype.UUID) ([]database.ChatMessage, error) {
	return r.q.GetChatMessagesByChatID(ctx, chatID)
}

func (r *ChatRepository) GetChatMessagesByIDs(ctx context.Context, ids []pgtype.UUID) ([]database.ChatMessage, error) {
	return r.q.GetChatMessagesByIDs(ctx, ids)
}

func (r *ChatRepository) GetChatMessagesByRole(ctx context.Context, params database.GetChatMessagesByRoleParams) ([]database.ChatMessage, error) {
	return r.q.GetChatMessagesByRole(ctx, params)
}

func (r *ChatRepository) GetChatMessagesBySender(ctx context.Context, params database.GetChatMessagesBySenderParams) ([]database.ChatMessage, error) {
	return r.q.GetChatMessagesBySender(ctx, params)
}

func (r *ChatRepository) GetChatMessagesWithSender(ctx context.Context, chatID pgtype.UUID) ([]database.GetChatMessagesWithSenderRow, error) {
	return r.q.GetChatMessagesWithSender(ctx, chatID)
}

func (r *ChatRepository) GetChatParticipant(ctx context.Context, params database.GetChatParticipantParams) (database.ChatParticipant, error) {
	return r.q.GetChatParticipant(ctx, params)
}

func (r *ChatRepository) GetChatParticipants(ctx context.Context, chatID pgtype.UUID) ([]database.ChatParticipant, error) {
	return r.q.GetChatParticipants(ctx, chatID)
}

func (r *ChatRepository) GetChatParticipantsWithUsers(ctx context.Context, chatID pgtype.UUID) ([]database.GetChatParticipantsWithUsersRow, error) {
	return r.q.GetChatParticipantsWithUsers(ctx, chatID)
}

func (r *ChatRepository) GetChatsByProjectAndType(ctx context.Context, params database.GetChatsByProjectAndTypeParams) ([]database.Chat, error) {
	return r.q.GetChatsByProjectAndType(ctx, params)
}

func (r *ChatRepository) GetChatsByProjectAndUser(ctx context.Context, params database.GetChatsByProjectAndUserParams) ([]database.Chat, error) {
	return r.q.GetChatsByProjectAndUser(ctx, params)
}

func (r *ChatRepository) GetChatsByProjectID(ctx context.Context, projectID pgtype.UUID) ([]database.Chat, error) {
	return r.q.GetChatsByProjectID(ctx, projectID)
}

func (r *ChatRepository) GetChatsByUserID(ctx context.Context, userID pgtype.UUID) ([]database.Chat, error) {
	return r.q.GetChatsByUserID(ctx, userID)
}

func (r *ChatRepository) GetLatestChatMessage(ctx context.Context, chatID pgtype.UUID) (database.ChatMessage, error) {
	return r.q.GetLatestChatMessage(ctx, chatID)
}

func (r *ChatRepository) GetMessageMentions(ctx context.Context, messageID pgtype.UUID) ([]database.GetMessageMentionsRow, error) {
	return r.q.GetMessageMentions(ctx, messageID)
}

func (r *ChatRepository) GetMessagesMentioningUser(ctx context.Context, mentionedUserID pgtype.UUID) ([]database.ChatMessage, error) {
	return r.q.GetMessagesMentioningUser(ctx, mentionedUserID)
}

func (r *ChatRepository) GetUserMentionsInChat(ctx context.Context, params database.GetUserMentionsInChatParams) ([]database.ChatMessage, error) {
	return r.q.GetUserMentionsInChat(ctx, params)
}

func (r *ChatRepository) RemoveChatParticipant(ctx context.Context, params database.RemoveChatParticipantParams) error {
	return r.q.RemoveChatParticipant(ctx, params)
}

func (r *ChatRepository) RenameChat(ctx context.Context, params database.RenameChatParams) (database.Chat, error) {
	return r.q.RenameChat(ctx, params)
}

func (r *ChatRepository) SearchChatMessages(ctx context.Context, params database.SearchChatMessagesParams) ([]database.ChatMessage, error) {
	return r.q.SearchChatMessages(ctx, params)
}

func (r *ChatRepository) SearchChatsByTitle(ctx context.Context, params database.SearchChatsByTitleParams) ([]database.Chat, error) {
	return r.q.SearchChatsByTitle(ctx, params)
}

func (r *ChatRepository) UpdateChatAIMetadata(ctx context.Context, params database.UpdateChatAIMetadataParams) (database.ChatAiMetadatum, error) {
	return r.q.UpdateChatAIMetadata(ctx, params)
}

func (r *ChatRepository) UpdateChatActivity(ctx context.Context, id pgtype.UUID) error {
	return r.q.UpdateChatActivity(ctx, id)
}

func (r *ChatRepository) UpdateChatMessage(ctx context.Context, params database.UpdateChatMessageParams) (database.ChatMessage, error) {
	return r.q.UpdateChatMessage(ctx, params)
}

func (r *ChatRepository) UpdateChatStatus(ctx context.Context, params database.UpdateChatStatusParams) (database.Chat, error) {
	return r.q.UpdateChatStatus(ctx, params)
}

func (r *ChatRepository) UpdateChatType(ctx context.Context, params database.UpdateChatTypeParams) (database.Chat, error) {
	return r.q.UpdateChatType(ctx, params)
}

func (r *ChatRepository) GetChatsByProjectIDs(ctx context.Context, projectIDs []pgtype.UUID) ([]database.Chat, error) {
	return r.q.GetChatsByProjectIDs(ctx, projectIDs)
}

func (r *ChatRepository) GetChatMessagesByChatIDs(ctx context.Context, chatIDs []pgtype.UUID) ([]database.ChatMessage, error) {
	return r.q.GetChatMessagesByChatIDs(ctx, chatIDs)
}

func (r *ChatRepository) GetChatParticipantsByChatIDs(ctx context.Context, chatIDs []pgtype.UUID) ([]database.ChatParticipant, error) {
	return r.q.GetChatParticipantsByChatIDs(ctx, chatIDs)
}
