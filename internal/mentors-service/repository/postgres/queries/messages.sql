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

-- name: ListMessagesByOwner :many
SELECT
    m.id,
    m.ticket_id,
    m.author_user_id,
    m.author_role,
    m.text,
    m.client_message_id,
    m.created_at
FROM mentors.messages m
JOIN mentors.tickets t ON m.ticket_id = t.id
WHERE t.hackathon_id = $1 AND t.owner_kind = $2 AND t.owner_id = $3
ORDER BY m.created_at ASC
LIMIT $4 OFFSET $5;
