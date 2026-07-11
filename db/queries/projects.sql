-- name: GetProjectsByOwnerID :many
SELECT *
FROM projects
WHERE owner_id = $1
  AND archived_at IS NULL;


-- name: GetProjectsForUser :many
SELECT DISTINCT p.*
FROM projects p
LEFT JOIN project_members pm
    ON p.id = pm.project_id
WHERE p.owner_id = $1
   OR pm.user_id = $1;


-- name: GetProjectByID :one
SELECT *
FROM projects
WHERE id = $1;


-- name: CreateProject :one
INSERT INTO projects (
    id,
    owner_id,
    name,
    description,
    visibility
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;


-- name: UpdateProject :one
UPDATE projects
SET
    name = $2,
    description = $3,
    visibility = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: ArchiveProject :exec
UPDATE projects
SET
    archived_at = NOW(),
    updated_at = NOW()
WHERE id = $1;


-- name: RestoreProject :exec
UPDATE projects
SET
    archived_at = NULL,
    updated_at = NOW()
WHERE id = $1;


-- name: DeleteProject :exec
DELETE
FROM projects
WHERE id = $1;


-- name: GetProjectMembers :many
SELECT *
FROM project_members
WHERE project_id = $1
ORDER BY created_at;


-- name: GetProjectMember :one
SELECT *
FROM project_members
WHERE project_id = $1
  AND user_id = $2;


-- name: AddProjectMember :one
INSERT INTO project_members (
    id,
    project_id,
    user_id,
    role
)
VALUES ($1, $2, $3, $4)
RETURNING *;


-- name: UpdateProjectMemberRole :one
UPDATE project_members
SET
    role = $3,
    updated_at = NOW()
WHERE project_id = $1
  AND user_id = $2
RETURNING *;


-- name: RemoveProjectMember :exec
DELETE
FROM project_members
WHERE project_id = $1
  AND user_id = $2;


-- name: GetArchivedProjectsByOwner :many
SELECT *
FROM projects
WHERE owner_id = $1
  AND archived_at IS NOT NULL;


-- name: GetArchivedProjectsForUser :many
SELECT DISTINCT p.*
FROM projects p
LEFT JOIN project_members pm
    ON p.id = pm.project_id
WHERE p.archived_at IS NOT NULL
  AND (
      p.owner_id = $1
      OR pm.user_id = $1
  );


-- name: TransferOwnership :one
UPDATE projects
SET
    owner_id = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetProjectByName :one
SELECT *
FROM projects
WHERE name = $1 AND owner_id = $2;