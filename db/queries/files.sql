-- name: CreateFile :one
INSERT INTO files (
    id,
    project_id,
    folder_id,
    name,
    size
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;


-- name: GetFileByID :one
SELECT *
FROM files
WHERE id = $1;


-- name: GetProjectFileByID :one
SELECT *
FROM files
WHERE id = $1
  AND project_id = $2;


-- name: GetFilesByProjectID :many
SELECT *
FROM files
WHERE project_id = $1
ORDER BY name;


-- name: GetFilesByFolderID :many
SELECT *
FROM files
WHERE folder_id IS NOT DISTINCT FROM $1
ORDER BY name;


-- name: GetRootFiles :many
SELECT *
FROM files
WHERE project_id = $1
  AND folder_id IS NULL
ORDER BY name;


-- name: GetFileByFolderAndName :one
SELECT *
FROM files
WHERE project_id = $1
  AND folder_id IS NOT DISTINCT FROM $2
  AND name = $3;


-- name: CheckFileNameExists :one
SELECT EXISTS (
    SELECT 1
    FROM files
    WHERE project_id = $1
      AND folder_id IS NOT DISTINCT FROM $2
      AND name = $3
);


-- name: SearchFiles :many
SELECT *
FROM files
WHERE project_id = $1
  AND name ILIKE '%' || $2 || '%'
ORDER BY name;


-- name: RenameFile :one
UPDATE files
SET
    name = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: MoveFile :one
UPDATE files
SET
    folder_id = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: UpdateFileSize :one
UPDATE files
SET
    size = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: DeleteFile :exec
DELETE
FROM files
WHERE id = $1;


-- name: CreateFolder :one
INSERT INTO folders (
    id,
    project_id,
    parent_folder_id,
    name
)
VALUES ($1, $2, $3, $4)
RETURNING *;


-- name: GetFolderByID :one
SELECT *
FROM folders
WHERE id = $1;


-- name: GetProjectFolderByID :one
SELECT *
FROM folders
WHERE id = $1
  AND project_id = $2;


-- name: GetFoldersByProjectID :many
SELECT *
FROM folders
WHERE project_id = $1
ORDER BY name;


-- name: GetChildFolders :many
SELECT *
FROM folders
WHERE parent_folder_id IS NOT DISTINCT FROM $1
ORDER BY name;


-- name: GetRootFolders :many
SELECT *
FROM folders
WHERE project_id = $1
  AND parent_folder_id IS NULL
ORDER BY name;


-- name: SearchFolders :many
SELECT *
FROM folders
WHERE project_id = $1
  AND name ILIKE '%' || $2 || '%'
ORDER BY name;


-- name: RenameFolder :one
UPDATE folders
SET
    name = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: MoveFolder :one
UPDATE folders
SET
    parent_folder_id = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: DeleteFolder :exec
DELETE
FROM folders
WHERE id = $1;


-- name: CheckFolderNameExists :one
SELECT EXISTS (
    SELECT 1
    FROM folders
    WHERE project_id = $1
      AND parent_folder_id IS NOT DISTINCT FROM $2
      AND name = $3
);


-- name: GetFolderContents :many
SELECT
    id,
    name,
    'folder' AS item_type,
    created_at
FROM folders
WHERE parent_folder_id IS NOT DISTINCT FROM $1
UNION ALL
SELECT
    id,
    name,
    'file' AS item_type,
    created_at
FROM files
WHERE folder_id IS NOT DISTINCT FROM $1
ORDER BY item_type, name;


-- name: CreateFileProperties :one
INSERT INTO file_properties (
    file_id,
    original_name,
    is_indexed,
    is_favorite,
    deleted_at
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;


-- name: GetFileProperties :one
SELECT *
FROM file_properties
WHERE file_id = $1;


-- name: UpdateFilePropertiesOriginalName :one
UPDATE file_properties
SET
    original_name = $2,
    updated_at = NOW()
WHERE file_id = $1
RETURNING *;


-- name: SetFavorite :exec
UPDATE file_properties
SET
    is_favorite = $2,
    updated_at = NOW()
WHERE file_id = $1;


-- name: MarkFileIndexed :exec
UPDATE file_properties
SET
    is_indexed = $2,
    updated_at = NOW()
WHERE file_id = $1;


-- name: SoftDeleteFile :exec
UPDATE file_properties
SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE file_id = $1;


-- name: RestoreFile :exec
UPDATE file_properties
SET
    deleted_at = NULL,
    updated_at = NOW()
WHERE file_id = $1;


-- name: DeleteFileProperties :exec
DELETE
FROM file_properties
WHERE file_id = $1;


-- name: CreateFileStorage :one
INSERT INTO file_storage (
    file_id,
    bucket_name,
    object_key,
    etag,
    version_id,
    checksum,
    mime_type,
    uploaded_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;


-- name: GetFileStorage :one
SELECT *
FROM file_storage
WHERE file_id = $1;


-- name: UpdateFileStorage :one
UPDATE file_storage
SET
    bucket_name = $2,
    object_key = $3,
    etag = $4,
    version_id = $5,
    checksum = $6,
    mime_type = $7,
    uploaded_by = $8,
    updated_at = NOW()
WHERE file_id = $1
RETURNING *;


-- name: DeleteFileStorage :exec
DELETE
FROM file_storage
WHERE file_id = $1;


-- name: CreateFileAIMetadata :one
INSERT INTO file_ai_metadata (
    file_id,
    extracted_text,
    embedding_model,
    embedding_synced,
    indexed_at
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;


-- name: GetFileAIMetadata :one
SELECT *
FROM file_ai_metadata
WHERE file_id = $1;


-- name: UpdateFileAIMetadata :one
UPDATE file_ai_metadata
SET
    extracted_text = $2,
    embedding_model = $3,
    embedding_synced = $4,
    indexed_at = $5,
    updated_at = NOW()
WHERE file_id = $1
RETURNING *;


-- name: DeleteFileAIMetadata :exec
DELETE
FROM file_ai_metadata
WHERE file_id = $1;


-- name: GetDeletedFiles :many
SELECT f.*
FROM files f
JOIN file_properties fp
ON f.id = fp.file_id
WHERE fp.deleted_at IS NOT NULL
  AND f.project_id = $1
ORDER BY fp.deleted_at DESC;


-- name: FileShare :one
INSERT INTO file_shares (
    id,
    file_id,
    shared_by,
    shared_with,
    permission
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;


-- name: GetFileShareByID :one
SELECT *
FROM file_shares
WHERE id = $1;


-- name: GetFileSharesByFileID :many
SELECT *
FROM file_shares
WHERE file_id = $1
ORDER BY created_at DESC;


-- name: GetFileSharesBySharedWith :many
SELECT *
FROM file_shares
WHERE shared_with = $1
ORDER BY created_at DESC;


-- name: GetFileShareByFileAndSharedWith :one
SELECT *
FROM file_shares
WHERE file_id = $1
  AND shared_with = $2;


-- name: UpdateFileSharePermission :one
UPDATE file_shares
SET
    permission = $2
WHERE id = $1
RETURNING *;


-- name: DeleteFileShare :exec
DELETE
FROM file_shares
WHERE id = $1;
