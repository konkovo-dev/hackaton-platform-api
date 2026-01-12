-- name: CreateRefreshToken :exec
INSERT INTO auth.refresh_tokens (id, user_id, token_hash, expires_at, revoked_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
FROM auth.refresh_tokens
WHERE token_hash = $1;

-- name: RevokeRefreshToken :exec
UPDATE auth.refresh_tokens
SET revoked_at = $2
WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: RevokeAllRefreshTokensByUserID :exec
UPDATE auth.refresh_tokens
SET revoked_at = $2
WHERE user_id = $1 AND revoked_at IS NULL;

