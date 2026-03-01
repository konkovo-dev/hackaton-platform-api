-- name: CreateOutboxEvent :exec
INSERT INTO team.outbox_events (
    id,
    aggregate_id,
    aggregate_type,
    event_type,
    payload,
    status,
    attempt_count,
    last_error,
    created_at,
    updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetPendingOutboxEvents :many
SELECT id, aggregate_id, aggregate_type, event_type, payload, status, attempt_count, last_error, created_at, updated_at
FROM team.outbox_events
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: UpdateOutboxEvent :exec
UPDATE team.outbox_events
SET status = $2,
    attempt_count = $3,
    last_error = $4,
    updated_at = $5
WHERE id = $1;
