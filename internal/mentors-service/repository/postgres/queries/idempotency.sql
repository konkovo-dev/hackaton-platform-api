-- name: GetIdempotencyKey :one
SELECT key, scope, request_hash, response_blob, created_at, expires_at
FROM mentors.idempotency_keys
WHERE key = $1 AND scope = $2 AND expires_at > NOW();

-- name: SetIdempotencyKey :execrows
INSERT INTO mentors.idempotency_keys (key, scope, request_hash, response_blob, created_at, expires_at)
VALUES ($1, $2, $3, $4, NOW(), $5)
ON CONFLICT (key, scope) DO NOTHING;
