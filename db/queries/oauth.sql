-- ============================================================
-- OAuth Accounts
-- ============================================================

-- name: GetOAuthAccount :one
SELECT
    id,
    user_id,
    provider,
    provider_user_id,
    created_at,
    updated_at
FROM oauth_accounts
WHERE provider = $1
  AND provider_user_id = $2
LIMIT 1;


-- name: GetOAuthAccountByUserAndProvider :one
SELECT
    id,
    user_id,
    provider,
    provider_user_id,
    created_at,
    updated_at
FROM oauth_accounts
WHERE user_id = $1
  AND provider = $2
LIMIT 1;


-- name: GetOAuthAccountsByUserID :many
SELECT
    id,
    user_id,
    provider,
    provider_user_id,
    created_at,
    updated_at
FROM oauth_accounts
WHERE user_id = $1
ORDER BY created_at ASC;


-- name: CreateOAuthAccount :one
INSERT INTO oauth_accounts (
    user_id,
    provider,
    provider_user_id
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING
    id,
    user_id,
    provider,
    provider_user_id,
    created_at,
    updated_at;


-- name: DeleteOAuthAccount :exec
DELETE FROM oauth_accounts
WHERE id = $1;


-- name: DeleteOAuthAccountByUserAndProvider :exec
DELETE FROM oauth_accounts
WHERE user_id = $1
  AND provider = $2;