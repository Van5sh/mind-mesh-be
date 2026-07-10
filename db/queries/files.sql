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
WHERE folder_id = $1
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
  AND folder_id = $2
  AND name = $3;


-- name: CheckFileNameExists :one
SELECT EXISTS (
    SELECT 1
    FROM files
    WHERE project_id = $1
      AND folder_id = $2
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
WHERE parent_folder_id = $1
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
      AND parent_folder_id = $2
      AND name = $3
);

-- name: GetFolderContents :many
SELECT
    id,
    name,
    'folder' AS item_type,
    created_at
FROM folders
WHERE parent_folder_id = $1
UNION ALL SELECT
    id,
    name,
    'file' AS item_type,
    created_at
FROM files
WHERE folder_id = $1

ORDER BY item_type, name;

-- name: GetFileProperties :one
SELECT *
FROM file_properties
WHERE file_id = $1;


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
    shared_with,
    permissions
)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetFileSharesByFileID :many
SELECT *
FROM file_shares
WHERE file_id = $1
ORDER BY created_at DESC;