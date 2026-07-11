-- name: CreateActivityLog :one
INSERT INTO activity_logs (
    id,
    project_id,
    user_id,
    action,
    entity_type,
    entity_id
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetActivityLogByID :one
SELECT *
FROM activity_logs
WHERE id = $1;

-- name: GetActivityLogsByProjectID :many
SELECT *
FROM activity_logs
WHERE project_id = $1
ORDER BY created_at DESC;

-- name: GetActivityLogsByUserID :many
SELECT *
FROM activity_logs
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetActivityLogsByAction :many
SELECT *
FROM activity_logs
WHERE action = $1
ORDER BY created_at DESC;

-- name: GetActivityLogsByEntity :many
SELECT *
FROM activity_logs
WHERE entity_type = $1
  AND entity_id = $2
ORDER BY created_at DESC;

-- name: GetActivityLogsByProjectAndUser :many
SELECT *
FROM activity_logs
WHERE project_id = $1
  AND user_id = $2
ORDER BY created_at DESC;

-- name: GetActivityLogsByProjectAndAction :many
SELECT *
FROM activity_logs
WHERE project_id = $1
  AND action = $2
ORDER BY created_at DESC;

-- name: GetActivityLogsByUserAndAction :many
SELECT *
FROM activity_logs
WHERE user_id = $1
  AND action = $2
ORDER BY created_at DESC;

-- name: GetActivityLogsByProjectUserAndAction :many
SELECT *
FROM activity_logs
WHERE project_id = $1
  AND user_id = $2
  AND action = $3
ORDER BY created_at DESC;

-- name: GetRecentActivityLogs :many
SELECT *
FROM activity_logs
ORDER BY created_at DESC
LIMIT $1;

-- name: CountActivityLogsByProjectID :one
SELECT COUNT(*) AS count
FROM activity_logs
WHERE project_id = $1;

-- name: CountActivityLogsByUserID :one
SELECT COUNT(*) AS count
FROM activity_logs
WHERE user_id = $1;


-- name: DeleteActivityLogByID :exec
DELETE
FROM activity_logs
WHERE id = $1;


-- name: DeleteActivityLogsByProjectID :exec
DELETE
FROM activity_logs
WHERE project_id = $1;