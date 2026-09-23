-- name: GetProjectsByOwnerID :many
SELECT *
FROM projects
WHERE owner_id = $1
  AND archived_at IS NULL;


-- name: GetProjectsForUser :many
-- Active projects only: archived ones are served by GetArchivedProjectsForUser
-- (the owner-scoped GetProjectsByOwnerID already excludes them the same way).
SELECT DISTINCT p.*
FROM projects p
LEFT JOIN project_members pm
    ON p.id = pm.project_id
WHERE p.archived_at IS NULL
  AND (
      p.owner_id = $1
      OR pm.user_id = $1
  );


-- name: GetProjectByID :one
SELECT *
FROM projects
WHERE id = $1;


-- name: CreateProject :one
INSERT INTO projects (
    owner_id,
    name,
    description,
    visibility
)
VALUES ($1, $2, $3, $4)
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
    project_id,
    user_id,
    role
)
VALUES ($1, $2, $3)
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

-- name: GetProjectMembersByProjectIDs :many
-- Batched form of GetProjectMembers, used by the Project.members dataloader.
SELECT *
FROM project_members
WHERE project_id = ANY($1::uuid[])
ORDER BY project_id, created_at;


-- name: GetProjectsByIDs :many
-- Batched project lookup, used by the Project scalar-field dataloader that
-- fills in "stub" projects (an object that only carries an ID).
SELECT *
FROM projects
WHERE id = ANY($1::uuid[]);


-- name: GetProjectsByOwnerIDs :many
-- Batched form of GetProjectsByOwnerID, used by the User.ownedProjects dataloader.
SELECT *
FROM projects
WHERE owner_id = ANY($1::uuid[])
  AND archived_at IS NULL
ORDER BY owner_id, created_at;


-- name: GetProjectMembersByUserIDs :many
-- Every membership row of many users, used by the User.projectMemberships
-- dataloader. (This replaces "list each user's projects, then look up the
-- membership row per project": a membership row is exactly what that produced.)
SELECT *
FROM project_members
WHERE user_id = ANY($1::uuid[])
ORDER BY user_id, created_at;
