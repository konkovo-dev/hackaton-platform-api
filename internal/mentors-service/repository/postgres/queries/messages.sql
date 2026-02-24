-- name: GetMessageByID :one
SELECT
    id,
    ticket_id,
    author_user_id,
    author_role,
    text,
    client_message_id,
    created_at
FROM mentors.messages
WHERE id = $1;

-- name: ListMessagesByTicket :many
SELECT
    id,
    ticket_id,
    author_user_id,
    author_role,
    text,
    client_message_id,
    created_at
FROM mentors.messages
WHERE ticket_id = $1
ORDER BY created_at ASC
LIMIT $2 OFFSET $3;

-- name: CreateMessage :one
INSERT INTO mentors.messages (
    id,
    ticket_id,
    author_user_id,
    author_role,
    text,
    client_message_id,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, NOW()
)
RETURNING
    id,
    ticket_id,
    author_user_id,
    author_role,
    text,
    client_message_id,
    created_at;

-- name: FindMessageByClientID :one
SELECT
    id,
    ticket_id,
    author_user_id,
    author_role,
    text,
    client_message_id,
    created_at
FROM mentors.messages
WHERE client_message_id = $1
LIMIT 1;
