-- name: CreateUser
INSERT INTO users (id,email, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id, username, email, created_at, updated_at;

-- name: GetUserByID
SELECT id, username, email, created_at, updated_at
FROM users WHERE id = $1;

-- name: GetUserByEmail
SELECT id, username, email, created_at, updated_at
FROM users WHERE email = $1;

-- name: UpdateUser
UPDATE users SET username = $2, email = $3, updated_at = NOW()
WHERE id = $1 RETURNING id, username, email, created_at, updated_at;

-- name: DeleteUser
DELETE FROM users WHERE id = $1;

-- name: UpdateUserPassword
UPDATE users SET password_hash = $2, updated_at = NOW()
WHERE id = $1 RETURNING id, username, email, created_at, updated_at;

-- name: CheckUserNameExists
SELECT EXISTS(
    SELECT 1
    FROM users
    WHERE username = $1
);