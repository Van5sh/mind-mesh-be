-- name: CreateChat :one
INSERT INTO chats (
    id,
    project_id,
    title,
    type
)
VALUES ($1, $2, $3, $4)
RETURNING *;


-- name: GetChatByID :one
SELECT *
FROM chats
WHERE id = $1;


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
    id,
    chat_id,
    sender_id,
    role,
    content
)
VALUES ($1, $2, $3, $4, $5)
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