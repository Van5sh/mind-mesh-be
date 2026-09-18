-- name: CreateFile :one
INSERT INTO files (
    folder_id,
    project_id,
    name,
    size
)
VALUES ($1, $2, $3,$4)
RETURNING *;

-- name: GetFileByID :one
SELECT *
FROM files
WHERE id = $1;

-- name: CreateProjectFile :one
INSERT INTO project_files (
    project_id,
    file_id,
    folder_id
)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetProjectFile :one
SELECT *
FROM project_files
WHERE project_id = $1
  AND file_id = $2;

-- name: GetProjectFileByID :one
SELECT f.*
FROM files f
WHERE f.id = $1
  AND f.project_id = $2;

-- name: CheckProjectContainsFile :one
SELECT EXISTS (
    SELECT 1
    FROM files
    WHERE project_id = $1
      AND id = $2
);

-- name: GetFilesByProjectID :many
SELECT f.*
FROM files f
WHERE f.project_id = $1
ORDER BY f.name;

-- name: GetFilesByProjectIDs :many
-- Batched form of GetFilesByProjectID, used by the Project.files dataloader
-- so resolving files for N projects is one query instead of N.
SELECT f.*
FROM files f
WHERE f.project_id = ANY($1::uuid[])
ORDER BY f.project_id, f.name;

-- name: CountProjectFiles :one
SELECT COUNT(*) AS count
FROM files
WHERE project_id = $1;

-- name: GetFilesByFolderID :many
SELECT *
FROM files
WHERE folder_id IS NOT DISTINCT FROM $1
ORDER BY name;

-- name: GetProjectFilesByFolderID :many
SELECT f.*
FROM files f
WHERE f.project_id = $1
  AND f.folder_id IS NOT DISTINCT FROM $2
ORDER BY f.name;

-- name: GetRootFiles :many
SELECT f.*
FROM files f
WHERE f.project_id = $1
  AND f.folder_id IS NULL
ORDER BY f.name;

-- name: GetFileByFolderAndName :one
SELECT *
FROM files
WHERE folder_id IS NOT DISTINCT FROM $1
  AND name = $2;

-- name: GetProjectFileByFolderAndName :one
SELECT f.*
FROM files f
WHERE f.project_id = $1
  AND f.folder_id IS NOT DISTINCT FROM $2
  AND f.name = $3;

-- name: CheckFileNameExists :one
SELECT EXISTS (
    SELECT 1
    FROM files
    WHERE folder_id IS NOT DISTINCT FROM $1
      AND name = $2
);

-- name: CheckProjectFileNameExists :one
SELECT EXISTS (
    SELECT 1
    FROM files f
    WHERE f.project_id = $1
      AND f.folder_id IS NOT DISTINCT FROM $2
      AND f.name = $3
);

-- name: SearchFiles :many
SELECT f.*
FROM files f
WHERE f.project_id = $1
  AND f.name ILIKE '%' || $2 || '%'
ORDER BY f.name;

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

-- name: MoveProjectFile :one
UPDATE project_files
SET
    folder_id = $3
WHERE project_id = $1
  AND file_id = $2
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

-- name: RemoveProjectFile :exec
DELETE
FROM project_files
WHERE project_id = $1
  AND file_id = $2;

-- name: CreateFolder :one
INSERT INTO folders (
    project_id,
    parent_folder_id,
    name
)
VALUES ($1, $2, $3)
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

-- name: GetStandaloneRootFolders :many
SELECT *
FROM folders
WHERE project_id IS NULL
  AND parent_folder_id IS NULL
ORDER BY name;


-- name: SearchFolders :many
SELECT *
FROM folders
WHERE project_id = $1
  AND name ILIKE '%' || $2 || '%'
ORDER BY name;

-- name: SearchStandaloneFolders :many
SELECT *
FROM folders
WHERE project_id IS NULL
  AND name ILIKE '%' || $1 || '%'
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
    f.id,
    f.name,
    'folder' AS item_type,
    f.created_at
FROM folders f
WHERE f.parent_folder_id IS NOT DISTINCT FROM $1
UNION ALL
SELECT
    fi.id,
    fi.name,
    'file' AS item_type,
    fi.created_at
FROM files fi
WHERE fi.folder_id IS NOT DISTINCT FROM $1
ORDER BY item_type, name;

-- name: GetProjectFolderContents :many
SELECT
    f.id,
    f.name,
    'folder' AS item_type,
    f.created_at
FROM folders f
WHERE f.project_id = $1
  AND f.parent_folder_id IS NOT DISTINCT FROM $2
UNION ALL
SELECT
    fi.id,
    fi.name,
    'file' AS item_type,
    fi.created_at
FROM files fi
WHERE fi.project_id = $1
  AND fi.folder_id IS NOT DISTINCT FROM $2
ORDER BY item_type, name;


-- name: CreateFileProperties :one
INSERT INTO file_properties (
    file_id,
    original_name,
    is_indexed,
    deleted_at
)
VALUES ($1, $2, $3, $4)
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


-- name: CreateUserFilePreference :one
INSERT INTO user_file_preferences (
    user_id,
    file_id,
    is_favorite
)
VALUES ($1, $2, $3)
RETURNING *;


-- name: GetUserFilePreference :one
SELECT *
FROM user_file_preferences
WHERE user_id = $1
  AND file_id = $2;


-- name: SetFavorite :one
INSERT INTO user_file_preferences (
    user_id,
    file_id,
    is_favorite
)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, file_id)
DO UPDATE SET
    is_favorite = EXCLUDED.is_favorite,
    updated_at = NOW()
RETURNING *;


-- name: DeleteUserFilePreference :exec
DELETE
FROM user_file_preferences
WHERE user_id = $1
  AND file_id = $2;


-- name: CreateMessageFileReference :one
INSERT INTO message_file_references (
    message_id,
    file_id
)
VALUES ($1, $2)
RETURNING *;


-- name: GetMessageFileReferences :many
SELECT f.*
FROM files f
JOIN message_file_references mfr
ON f.id = mfr.file_id
WHERE mfr.message_id = $1
ORDER BY f.name;


-- name: DeleteMessageFileReferences :exec
DELETE
FROM message_file_references
WHERE message_id = $1;


-- name: GetMessagesReferencingFile :many
SELECT cm.*
FROM chat_messages cm
JOIN message_file_references mfr
ON cm.id = mfr.message_id
WHERE mfr.file_id = $1
ORDER BY cm.created_at ASC;


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


-- name: GetFilesPendingEmbedding :many
SELECT f.*
FROM files f
JOIN file_ai_metadata fam
ON f.id = fam.file_id
WHERE fam.embedding_synced = FALSE
ORDER BY f.created_at ASC;


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


-- name: MarkFileEmbeddingSynced :one
UPDATE file_ai_metadata
SET
    embedding_synced = $2,
    indexed_at = $3,
    updated_at = NOW()
WHERE file_id = $1
RETURNING *;


-- name: MarkFileProcessingStarted :one
UPDATE file_ai_metadata
SET
    processing_status = 'PROCESSING',
    error_message = NULL,
    updated_at = NOW()
WHERE file_id = $1
RETURNING *;


-- name: CompleteFileProcessing :one
UPDATE file_ai_metadata
SET
    processing_status = 'COMPLETED',
    summary = $2,
    error_message = NULL,
    embedding_synced = TRUE,
    indexed_at = NOW(),
    updated_at = NOW()
WHERE file_id = $1
RETURNING *;


-- name: FailFileProcessing :one
UPDATE file_ai_metadata
SET
    processing_status = 'FAILED',
    error_message = $2,
    updated_at = NOW()
WHERE file_id = $1
RETURNING *;


-- name: GetFilesByProcessingStatus :many
SELECT f.*
FROM files f
JOIN file_ai_metadata fam
ON f.id = fam.file_id
WHERE f.project_id = $1
  AND fam.processing_status = $2
ORDER BY f.created_at ASC;


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
    file_id,
    shared_by,
    shared_with,
    permission
)
VALUES ($1, $2, $3, $4)
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


-- name: GetFilesSharedByUser :many
SELECT *
FROM file_shares
WHERE shared_by = $1
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


-- name: GetFavoritesFiles :many
SELECT f.*
FROM files f
JOIN user_file_preferences ufp
ON f.id = ufp.file_id
WHERE ufp.user_id = $1
  AND ufp.is_favorite = TRUE
  AND f.project_id = $2;


-- name: GetIndexedFiles :many
SELECT f.*
FROM files f
JOIN file_properties fp
ON f.id = fp.file_id
WHERE fp.is_indexed = TRUE
  AND f.project_id = $1;


-- name: GetFilesByIDs :many
SELECT *
FROM files
WHERE id = ANY($1::UUID[])
ORDER BY name;


-- name: UpdateFileProperties :one
UPDATE file_properties
SET
    original_name = $2,
    is_indexed = $3,
    deleted_at = $4,
    updated_at = NOW()
WHERE file_id = $1
RETURNING *;

-- name: FolderNameExists :one
SELECT EXISTS (
    SELECT 1
    FROM folders
    WHERE name = $1
      AND parent_folder_id IS NOT DISTINCT FROM $2
      AND project_id IS NOT DISTINCT FROM $3
);

-- name: FileNameExistsInFolder :one
SELECT EXISTS (
    SELECT 1
    FROM files
    WHERE project_id = $1
      AND folder_id IS NOT DISTINCT FROM $2
      AND name = $3
);