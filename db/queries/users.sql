-- ============================================================
-- Users
-- ============================================================

-- name: CreateUser :one
INSERT INTO users (
    id,
    username,
    email
)
VALUES ($1, $2, $3)
RETURNING *;


-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;


-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;


-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1;


-- name: GetUsersByIDs :many
SELECT *
FROM users
WHERE id = ANY($1::UUID[])
ORDER BY username;


-- name: GetAllUsers :many
SELECT *
FROM users
ORDER BY username;


-- name: UpdateUser :one
UPDATE users
SET
    username = $2,
    email = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;


-- ============================================================
-- User Existence Checks
-- ============================================================

-- name: CheckUsernameExists :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE username = $1
);


-- name: CheckEmailExists :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE email = $1
);


-- ============================================================
-- User Profiles
-- ============================================================

-- name: CreateUserProfile :one
INSERT INTO user_profiles (
    user_id,
    first_name,
    last_name,
    bio,
    avatar_url
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;


-- name: GetUserProfile :one
SELECT *
FROM user_profiles
WHERE user_id = $1;


-- name: GetUserWithProfile :one
SELECT
    u.id,
    u.username,
    u.email,
    u.created_at,
    u.updated_at,
    p.first_name,
    p.last_name,
    p.bio,
    p.avatar_url
FROM users u
LEFT JOIN user_profiles p
    ON u.id = p.user_id
WHERE u.id = $1;


-- name: UpdateUserProfile :one
UPDATE user_profiles
SET
    first_name = $2,
    last_name = $3,
    bio = $4,
    avatar_url = $5,
    updated_at = NOW()
WHERE user_id = $1
RETURNING *;


-- name: UpdateUserAvatar :one
UPDATE user_profiles
SET
    avatar_url = $2,
    updated_at = NOW()
WHERE user_id = $1
RETURNING *;


-- name: DeleteUserProfile :exec
DELETE FROM user_profiles
WHERE user_id = $1;