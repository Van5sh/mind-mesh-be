-- ============================================================
-- Sessions
-- ============================================================

-- name: CreateSession :one
INSERT INTO sessions (
    user_id,
    expires_at
)
VALUES (
    $1,
    $2
)
RETURNING
    id,
    user_id,
    expires_at,
    created_at;


-- name: GetSessionByID :one
SELECT
    id,
    user_id,
    expires_at,
    created_at
FROM sessions
WHERE id = $1
LIMIT 1;


-- name: GetValidSession :one
SELECT
    id,
    user_id,
    expires_at,
    created_at
FROM sessions
WHERE id = $1
  AND expires_at > NOW()
LIMIT 1;


-- name: GetSessionsByUserID :many
SELECT
    id,
    user_id,
    expires_at,
    created_at
FROM sessions
WHERE user_id = $1
ORDER BY created_at DESC;


-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;


-- name: DeleteSessionsByUserID :exec
DELETE FROM sessions
WHERE user_id = $1;


-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at <= NOW();