-- name: GetFlowchartsByProjectID :many
SELECT *
FROM flowcharts
WHERE project_id = $1
ORDER BY created_at DESC;


-- name: GetFlowchartByID :one
SELECT *
FROM flowcharts
WHERE id = $1;


-- name: GetLatestFlowchart :one
SELECT *
FROM flowcharts
WHERE project_id = $1
ORDER BY created_at DESC
LIMIT 1;


-- name: CreateFlowchart :one
INSERT INTO flowcharts (
    project_id,
    name,
    data,
    generated_by,
    generated_by_ai,
    status,
    source_chat_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;


-- name: UpdateFlowchart :one
UPDATE flowcharts
SET
    name = $2,
    data = $3,
    generated_by = $4,
    generated_by_ai = $5,
    status = $6,
    source_chat_id = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: UpdateFlowchartStatus :one
UPDATE flowcharts
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: RenameFlowchart :one
UPDATE flowcharts
SET
    name = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: GetFlowchartsByChatID :many
SELECT *
FROM flowcharts
WHERE source_chat_id = $1
ORDER BY created_at DESC;


-- name: GetFlowchartsByGenerator :many
SELECT *
FROM flowcharts
WHERE generated_by = $1
ORDER BY created_at DESC;


-- name: DeleteFlowchart :exec
DELETE
FROM flowcharts
WHERE id = $1;


-- name: GetFlowchartWithProject :one
SELECT
    f.*,
    p.name AS project_name,
    p.owner_id,
    p.visibility
FROM flowcharts f
JOIN projects p
ON f.project_id = p.id
WHERE f.id = $1;

-- name: GetAIFlowCharts :many
SELECT *
FROM flowcharts
WHERE generated_by_ai = TRUE
ORDER BY created_at DESC;

-- name: GetFlowchartsByStatus :many
SELECT *
FROM flowcharts
WHERE status = $1
ORDER BY created_at DESC;

-- name: GetFlowchartsByProjectAndStatus :many
SELECT *
FROM flowcharts
WHERE project_id = $1
    AND status = $2
ORDER BY created_at DESC;


-- name: GetFlowchartsByProjectIDs :many
-- Batched form of GetFlowchartsByProjectID, used by the Project.flowcharts dataloader.
SELECT *
FROM flowcharts
WHERE project_id = ANY($1::uuid[])
ORDER BY project_id, created_at DESC;
