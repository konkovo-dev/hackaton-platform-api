-- name: CreateCredentials :exec
INSERT INTO auth.credentials (user_id, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4);

-- name: GetCredentialsByUserID :one
SELECT user_id, password_hash, created_at, updated_at
FROM auth.credentials
WHERE user_id = $1;

-- name: UpdateCredentials :exec
UPDATE auth.credentials
SET password_hash = $2, updated_at = $3
WHERE user_id = $1;

