-- name: CreateChat :one
INSERT INTO chats (
    project_id,
    title,
    type
)
VALUES ($1, $2, $3)
RETURNING *;


-- name: GetChatByID :one
SELECT *
FROM chats
WHERE id = $1;

-- name: CreateChatParticipant :one
INSERT INTO chat_participants (
    chat_id,
    user_id
)
VALUES ($1, $2)
RETURNING *;

-- name: GetChatParticipant :one
SELECT *
FROM chat_participants
WHERE chat_id = $1
  AND user_id = $2;

-- name: GetChatParticipants :many
SELECT *
FROM chat_participants
WHERE chat_id = $1
ORDER BY joined_at ASC;

-- name: GetChatParticipantsWithUsers :many
SELECT
    cp.chat_id,
    cp.user_id,
    cp.joined_at,
    u.username,
    u.email
FROM chat_participants cp
JOIN users u
ON cp.user_id = u.id
WHERE cp.chat_id = $1
ORDER BY cp.joined_at ASC;

-- name: GetChatsByUserID :many
SELECT c.*
FROM chats c
JOIN chat_participants cp
ON c.id = cp.chat_id
WHERE cp.user_id = $1
ORDER BY c.last_activity_at DESC;

-- name: GetChatsByProjectAndUser :many
SELECT c.*
FROM chats c
JOIN chat_participants cp
ON c.id = cp.chat_id
WHERE c.project_id = $1
  AND cp.user_id = $2
ORDER BY c.last_activity_at DESC;

-- name: CheckUserInChat :one
SELECT EXISTS (
    SELECT 1
    FROM chat_participants
    WHERE chat_id = $1
      AND user_id = $2
);

-- name: RemoveChatParticipant :exec
DELETE
FROM chat_participants
WHERE chat_id = $1
  AND user_id = $2;


-- name: GetChatsByProjectID :many
SELECT *
FROM chats
WHERE project_id = $1
ORDER BY last_activity_at DESC;


-- name: GetChatsByProjectAndType :many
SELECT *
FROM chats
WHERE project_id = $1
  AND type = $2
ORDER BY last_activity_at DESC;


-- name: SearchChatsByTitle :many
SELECT *
FROM chats
WHERE project_id = $1
  AND title ILIKE '%' || $2 || '%'
ORDER BY last_activity_at DESC;


-- name: RenameChat :one
UPDATE chats
SET
    title = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: UpdateChatStatus :one
UPDATE chats
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: UpdateChatType :one
UPDATE chats
SET
    type = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: UpdateChatActivity :exec
UPDATE chats
SET
    last_activity_at = NOW()
WHERE id = $1;


-- name: DeleteChat :exec
DELETE
FROM chats
WHERE id = $1;


-- name: CreateChatMessage :one
INSERT INTO chat_messages (
    chat_id,
    sender_id,
    role,
    content
)
VALUES ($1, $2, $3, $4)
RETURNING *;


-- name: GetChatMessageByID :one
SELECT *
FROM chat_messages
WHERE id = $1;


-- name: GetChatMessagesByChatID :many
SELECT *
FROM chat_messages
WHERE chat_id = $1
ORDER BY created_at ASC;


-- name: GetChatMessagesWithSender :many
SELECT
    cm.*,
    u.username
FROM chat_messages cm
LEFT JOIN users u
ON cm.sender_id = u.id
WHERE cm.chat_id = $1
ORDER BY cm.created_at ASC;


-- name: GetLatestChatMessage :one
SELECT *
FROM chat_messages
WHERE chat_id = $1
ORDER BY created_at DESC
LIMIT 1;


-- name: GetChatMessagesBySender :many
SELECT *
FROM chat_messages
WHERE chat_id = $1
  AND sender_id = $2
ORDER BY created_at ASC;


-- name: GetChatMessagesByRole :many
SELECT *
FROM chat_messages
WHERE chat_id = $1
  AND role = $2
ORDER BY created_at ASC;


-- name: GetChatMessagesByIDs :many
SELECT *
FROM chat_messages
WHERE id = ANY($1::UUID[])
ORDER BY created_at ASC;


-- name: SearchChatMessages :many
SELECT cm.*
FROM chat_messages cm
JOIN chats c
ON cm.chat_id = c.id
WHERE c.project_id = $1
  AND cm.content ILIKE '%' || $2 || '%'
ORDER BY cm.created_at DESC;


-- name: UpdateChatMessage :one
UPDATE chat_messages
SET
    content = $2
WHERE id = $1
RETURNING *;


-- name: DeleteChatMessage :exec
DELETE
FROM chat_messages
WHERE id = $1;

-- name: CreateMessageMention :one
INSERT INTO message_mentions (
    message_id,
    mentioned_user_id
)
VALUES ($1, $2)
RETURNING *;


-- name: GetMessageMentions :many
SELECT
    mm.message_id,
    mm.mentioned_user_id,
    mm.created_at,
    u.username,
    u.email
FROM message_mentions mm
JOIN users u
ON mm.mentioned_user_id = u.id
WHERE mm.message_id = $1
ORDER BY mm.created_at ASC;


-- name: DeleteMessageMentions :exec
DELETE
FROM message_mentions
WHERE message_id = $1;


-- name: GetMessagesMentioningUser :many
SELECT cm.*
FROM chat_messages cm
JOIN message_mentions mm
ON cm.id = mm.message_id
WHERE mm.mentioned_user_id = $1
ORDER BY cm.created_at DESC;


-- name: GetUserMentionsInChat :many
SELECT cm.*
FROM chat_messages cm
JOIN message_mentions mm
ON cm.id = mm.message_id
WHERE cm.chat_id = $1
  AND mm.mentioned_user_id = $2
ORDER BY cm.created_at DESC;


-- name: CreateChatAIMetadata :one
INSERT INTO chat_ai_metadata (
    message_id,
    embedding_model,
    embedding_synced,
    indexed_at
)
VALUES ($1, $2, $3, $4)
RETURNING *;


-- name: GetChatAIMetadata :one
SELECT *
FROM chat_ai_metadata
WHERE message_id = $1;


-- name: UpdateChatAIMetadata :one
UPDATE chat_ai_metadata
SET
    embedding_model = $2,
    embedding_synced = $3,
    indexed_at = $4
WHERE message_id = $1
RETURNING *;


-- name: DeleteChatAIMetadata :exec
DELETE
FROM chat_ai_metadata
WHERE message_id = $1;

-- name: GeneratingChats :many
SELECT *
FROM chats
WHERE project_id = $1
    AND status = 'GENERATING';

-- name: GetArchivedChats :many
SELECT *
FROM chats
WHERE project_id = $1
    AND status = 'ARCHIVED'
ORDER BY last_activity_at DESC;

-- name: GetActiveChats :many
SELECT *
FROM chats
WHERE project_id = $1
    AND status = 'ACTIVE'
ORDER BY last_activity_at DESC;

-- name: GetArchivedChatMessages :many
SELECT cm.*
FROM chat_messages cm
JOIN chats c
ON cm.chat_id = c.id
WHERE cm.chat_id = $1
    AND c.status = 'ARCHIVED'
ORDER BY cm.created_at ASC;
