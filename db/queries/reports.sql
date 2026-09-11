-- name: CreateReport :one
WITH new_report AS (
    INSERT INTO reports (
        project_id,
        title,
        content,
        format
    )
    VALUES ($1, $2, $3, $4)
    RETURNING *
),
new_properties AS (
    INSERT INTO report_properties (
        report_id,
        generated_by,
        generated_by_ai,
        status,
        source_chat_id
    )
    SELECT id, $5, $6, $7, $8
    FROM new_report
    RETURNING *
)
SELECT
    r.id,
    r.project_id,
    r.title,
    r.content,
    r.format,
    r.created_at,
    r.updated_at,
    rp.generated_by,
    rp.generated_by_ai,
    rp.status,
    rp.source_chat_id
FROM new_report r
JOIN new_properties rp
ON r.id = rp.report_id;


-- name: GetReportByID :one
SELECT
    r.id,
    r.project_id,
    r.title,
    r.content,
    r.format,
    r.created_at,
    r.updated_at,
    rp.generated_by,
    rp.generated_by_ai,
    rp.status,
    rp.source_chat_id
FROM reports r
JOIN report_properties rp
ON r.id = rp.report_id
WHERE r.id = $1;


-- name: GetReportsByProjectID :many
SELECT
    r.id,
    r.project_id,
    r.title,
    r.content,
    r.format,
    r.created_at,
    r.updated_at,
    rp.generated_by,
    rp.generated_by_ai,
    rp.status,
    rp.source_chat_id
FROM reports r
JOIN report_properties rp
ON r.id = rp.report_id
WHERE r.project_id = $1
ORDER BY r.created_at DESC;


-- name: GetReportsByChatID :many
SELECT
    r.id,
    r.project_id,
    r.title,
    r.content,
    r.format,
    r.created_at,
    r.updated_at,
    rp.generated_by,
    rp.generated_by_ai,
    rp.status,
    rp.source_chat_id
FROM reports r
JOIN report_properties rp
ON r.id = rp.report_id
WHERE rp.source_chat_id = $1
ORDER BY r.created_at DESC;


-- name: GetReportsByGenerator :many
SELECT
    r.id,
    r.project_id,
    r.title,
    r.content,
    r.format,
    r.created_at,
    r.updated_at,
    rp.generated_by,
    rp.generated_by_ai,
    rp.status,
    rp.source_chat_id
FROM reports r
JOIN report_properties rp
ON r.id = rp.report_id
WHERE rp.generated_by = $1
ORDER BY r.created_at DESC;


-- name: GetAIReports :many
SELECT
    r.id,
    r.project_id,
    r.title,
    r.content,
    r.format,
    r.created_at,
    r.updated_at,
    rp.generated_by,
    rp.generated_by_ai,
    rp.status,
    rp.source_chat_id
FROM reports r
JOIN report_properties rp
ON r.id = rp.report_id
WHERE r.project_id = $1
  AND rp.generated_by_ai = TRUE
ORDER BY r.created_at DESC;


-- name: GetReportsByStatus :many
SELECT
    r.id,
    r.project_id,
    r.title,
    r.content,
    r.format,
    r.created_at,
    r.updated_at,
    rp.generated_by,
    rp.generated_by_ai,
    rp.status,
    rp.source_chat_id
FROM reports r
JOIN report_properties rp
ON r.id = rp.report_id
WHERE r.project_id = $1
  AND rp.status = $2
ORDER BY r.created_at DESC;


-- name: GetReportsByFormat :many
SELECT
    r.id,
    r.project_id,
    r.title,
    r.content,
    r.format,
    r.created_at,
    r.updated_at,
    rp.generated_by,
    rp.generated_by_ai,
    rp.status,
    rp.source_chat_id
FROM reports r
JOIN report_properties rp
ON r.id = rp.report_id
WHERE r.project_id = $1
  AND r.format = $2
ORDER BY r.created_at DESC;


-- name: UpdateReport :one
WITH updated_report AS (
    UPDATE reports
    SET
        title = $2,
        content = $3,
        format = $4,
        updated_at = NOW()
    WHERE id = $1
    RETURNING *
)
SELECT
    r.id,
    r.project_id,
    r.title,
    r.content,
    r.format,
    r.created_at,
    r.updated_at,
    rp.generated_by,
    rp.generated_by_ai,
    rp.status,
    rp.source_chat_id
FROM updated_report r
JOIN report_properties rp
ON r.id = rp.report_id;


-- name: UpdateReportStatus :one
WITH updated_properties AS (
    UPDATE report_properties
    SET
        status = $2
    WHERE report_id = $1
    RETURNING *
)
SELECT
    r.id,
    r.project_id,
    r.title,
    r.content,
    r.format,
    r.created_at,
    r.updated_at,
    rp.generated_by,
    rp.generated_by_ai,
    rp.status,
    rp.source_chat_id
FROM reports r
JOIN updated_properties rp
ON r.id = rp.report_id
WHERE r.id = $1;


-- name: DeleteReport :exec
DELETE
FROM reports
WHERE id = $1;


-- name: GetReportWithProject :one
SELECT
    r.id,
    r.project_id,
    r.title,
    r.content,
    r.format,
    r.created_at,
    r.updated_at,
    rp.generated_by,
    rp.generated_by_ai,
    rp.status,
    rp.source_chat_id,
    p.name AS project_name,
    p.owner_id,
    p.visibility
FROM reports r
JOIN report_properties rp
ON r.id = rp.report_id
JOIN projects p
ON r.project_id = p.id
WHERE r.id = $1;
