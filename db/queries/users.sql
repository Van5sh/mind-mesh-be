-- name: CreateUser
INSERT INTO users (id, email, password_hash, name, role, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW());

-- name: GetUserByEmail
SELECT id, email, password_hash, name, role, created_at, updated_at
FROM users WHERE email = $1;

-- name: GetUserByID
SELECT id, email, password_hash, name, role, created_at, updated_at
FROM users WHERE id = $1;

-- name: UpdateUserPassword
UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1;

-- name: UpdateUser
UPDATE users SET email = $2, password_hash = $3, name = $4, role = $5, updated_at = NOW() WHERE id = $1;

-- name: DeleteUser
DELETE FROM users WHERE id = $1;