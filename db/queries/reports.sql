-- name: CreateReport :one
INSERT INTO reports (
    id,
    project_id,
    title,
    content,
    generated_by,
    generated_by_ai,
    status,
    source_chat_id,
    format
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;


-- name: GetReportByID :one
SELECT *
FROM reports
WHERE id = $1;


-- name: GetReportsByProjectID :many
SELECT *
FROM reports
WHERE project_id = $1
ORDER BY created_at DESC;


-- name: GetReportsByChatID :many
SELECT *
FROM reports
WHERE source_chat_id = $1
ORDER BY created_at DESC;


-- name: GetReportsByGenerator :many
SELECT *
FROM reports
WHERE generated_by = $1
ORDER BY created_at DESC;


-- name: GetAIReports :many
SELECT *
FROM reports
WHERE project_id = $1
  AND generated_by_ai = TRUE
ORDER BY created_at DESC;


-- name: UpdateReport :one
UPDATE reports
SET
    title = $2,
    content = $3,
    format = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: UpdateReportStatus :one
UPDATE reports
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: DeleteReport :exec
DELETE
FROM reports
WHERE id = $1;


-- name: GetReportWithProject :one
SELECT
    r.*,
    p.name AS project_name,
    p.owner_id,
    p.visibility
FROM reports r
JOIN projects p
ON r.project_id = p.id
WHERE r.id = $1;