-- name: ListTeams :many
SELECT 
    id,
    hackathon_id,
    name,
    description,
    is_joinable,
    created_at,
    updated_at
FROM team.teams
WHERE hackathon_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetTeamByID :one
SELECT
    id,
    hackathon_id,
    name,
    description,
    is_joinable,
    created_at,
    updated_at
FROM team.teams
WHERE id = $1;

-- name: GetTeamByIDAndHackathonID :one
SELECT
    id,
    hackathon_id,
    name,
    description,
    is_joinable,
    created_at,
    updated_at
FROM team.teams
WHERE id = $1 AND hackathon_id = $2;

-- name: CreateTeam :one
INSERT INTO team.teams (
    id,
    hackathon_id,
    name,
    description,
    is_joinable,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, NOW(), NOW()
)
RETURNING
    id,
    hackathon_id,
    name,
    description,
    is_joinable,
    created_at,
    updated_at;

-- name: CheckTeamNameExists :one
SELECT EXISTS(
    SELECT 1
    FROM team.teams
    WHERE hackathon_id = $1 AND name = $2
);

-- name: DeleteTeam :exec
DELETE FROM team.teams
WHERE id = $1;

-- name: UpdateTeam :one
UPDATE team.teams
SET
    name = $2,
    description = $3,
    is_joinable = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    hackathon_id,
    name,
    description,
    is_joinable,
    created_at,
    updated_at;
