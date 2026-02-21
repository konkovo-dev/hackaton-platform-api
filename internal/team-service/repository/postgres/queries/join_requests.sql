-- name: CancelCompetingJoinRequests :exec
UPDATE team.join_requests
SET
    status = 'canceled',
    updated_at = NOW()
WHERE requester_user_id = $1
  AND hackathon_id = $2
  AND status = 'pending';

-- name: ListJoinRequests :many
SELECT
    id,
    hackathon_id,
    team_id,
    vacancy_id,
    requester_user_id,
    message,
    status,
    created_at,
    updated_at,
    expires_at
FROM team.join_requests
WHERE team_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountJoinRequests :one
SELECT COUNT(*)
FROM team.join_requests
WHERE team_id = $1;

-- name: CreateJoinRequest :exec
INSERT INTO team.join_requests (
    id,
    hackathon_id,
    team_id,
    vacancy_id,
    requester_user_id,
    message,
    status,
    created_at,
    updated_at,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);

-- name: ListMyJoinRequests :many
SELECT
    id,
    hackathon_id,
    team_id,
    vacancy_id,
    requester_user_id,
    message,
    status,
    created_at,
    updated_at,
    expires_at
FROM team.join_requests
WHERE requester_user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountMyJoinRequests :one
SELECT COUNT(*)
FROM team.join_requests
WHERE requester_user_id = $1;

-- name: GetJoinRequestByID :one
SELECT
    id,
    hackathon_id,
    team_id,
    vacancy_id,
    requester_user_id,
    message,
    status,
    created_at,
    updated_at,
    expires_at
FROM team.join_requests
WHERE id = $1;

-- name: UpdateJoinRequestStatus :one
UPDATE team.join_requests
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    hackathon_id,
    team_id,
    vacancy_id,
    requester_user_id,
    message,
    status,
    created_at,
    updated_at,
    expires_at;
