-- name: GetIdempotencyKey :one
SELECT
    key,
    scope,
    request_hash,
    response_blob,
    created_at,
    expires_at
FROM judging.idempotency_keys
WHERE key = $1 AND scope = $2 AND expires_at > NOW();

-- name: SetIdempotencyKey :exec
INSERT INTO judging.idempotency_keys (
    key,
    scope,
    request_hash,
    response_blob,
    created_at,
    expires_at
) VALUES (
    $1, $2, $3, $4, NOW(), $5
)
ON CONFLICT (key, scope) DO UPDATE
SET
    request_hash = EXCLUDED.request_hash,
    response_blob = EXCLUDED.response_blob,
    expires_at = EXCLUDED.expires_at;
