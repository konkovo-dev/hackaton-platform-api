-- name: GetTicketByID :one
SELECT
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    status,
    assigned_mentor_user_id,
    created_at,
    updated_at,
    closed_at
FROM mentors.tickets
WHERE id = $1;

-- name: GetTicketByIDAndHackathonID :one
SELECT
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    status,
    assigned_mentor_user_id,
    created_at,
    updated_at,
    closed_at
FROM mentors.tickets
WHERE id = $1 AND hackathon_id = $2;

-- name: ListTicketsByOwner :many
SELECT
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    status,
    assigned_mentor_user_id,
    created_at,
    updated_at,
    closed_at
FROM mentors.tickets
WHERE hackathon_id = $1 AND owner_kind = $2 AND owner_id = $3
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: FindOpenTicketByOwner :one
SELECT
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    status,
    assigned_mentor_user_id,
    created_at,
    updated_at,
    closed_at
FROM mentors.tickets
WHERE hackathon_id = $1 AND owner_kind = $2 AND owner_id = $3 AND status = 'open'
LIMIT 1
FOR UPDATE;

-- name: ListTicketsByMentor :many
SELECT
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    status,
    assigned_mentor_user_id,
    created_at,
    updated_at,
    closed_at
FROM mentors.tickets
WHERE hackathon_id = $1 AND assigned_mentor_user_id = $2
ORDER BY updated_at DESC
LIMIT $3 OFFSET $4;

-- name: ListAllTickets :many
SELECT
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    status,
    assigned_mentor_user_id,
    created_at,
    updated_at,
    closed_at
FROM mentors.tickets
WHERE hackathon_id = $1
ORDER BY updated_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateTicket :one
INSERT INTO mentors.tickets (
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    status,
    assigned_mentor_user_id,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, NOW(), NOW()
)
RETURNING
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    status,
    assigned_mentor_user_id,
    created_at,
    updated_at,
    closed_at;

-- name: CreateOrGetOpenTicket :one
INSERT INTO mentors.tickets (
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    status,
    assigned_mentor_user_id,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, 'open', NULL, NOW(), NOW()
)
ON CONFLICT (hackathon_id, owner_kind, owner_id) WHERE status = 'open'
DO UPDATE SET updated_at = mentors.tickets.updated_at
RETURNING
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    status,
    assigned_mentor_user_id,
    created_at,
    updated_at,
    closed_at;

-- name: UpdateTicketStatus :one
UPDATE mentors.tickets
SET
    status = $2,
    closed_at = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    status,
    assigned_mentor_user_id,
    created_at,
    updated_at,
    closed_at;

-- name: CountOpenTicketsByMentor :one
SELECT COUNT(*)
FROM mentors.tickets
WHERE hackathon_id = $1 AND assigned_mentor_user_id = $2 AND status = 'open';

-- name: ClaimTicket :execrows
UPDATE mentors.tickets
SET
    assigned_mentor_user_id = $2,
    updated_at = NOW()
WHERE id = $1
  AND status = 'open'
  AND assigned_mentor_user_id IS NULL;

-- name: AssignTicket :execrows
UPDATE mentors.tickets
SET
    assigned_mentor_user_id = $2,
    updated_at = NOW()
WHERE id = $1
  AND status = 'open';
