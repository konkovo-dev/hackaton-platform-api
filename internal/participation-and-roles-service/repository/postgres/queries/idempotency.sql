-- name: GetIdempotencyKey :one
SELECT key, scope, request_hash, response_blob, created_at, expires_at
FROM participation_and_roles.idempotency_keys
WHERE key = $1 AND scope = $2 AND expires_at > NOW();

-- name: CreateIdempotencyKey :exec
INSERT INTO participation_and_roles.idempotency_keys (
    key,
    scope,
    request_hash,
    response_blob,
    created_at,
    expires_at
) VALUES (
    $1, $2, $3, $4, NOW(), $5
);

-- name: DeleteExpiredIdempotencyKeys :exec
DELETE FROM participation_and_roles.idempotency_keys
WHERE expires_at <= NOW();

