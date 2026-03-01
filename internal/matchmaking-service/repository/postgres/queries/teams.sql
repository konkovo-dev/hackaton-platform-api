-- name: UpsertTeam :exec
INSERT INTO matchmaking.teams (
    team_id,
    hackathon_id,
    name,
    description,
    is_joinable,
    updated_at
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (team_id) DO UPDATE SET
    hackathon_id = EXCLUDED.hackathon_id,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_joinable = EXCLUDED.is_joinable,
    updated_at = EXCLUDED.updated_at;

-- name: GetTeamByID :one
SELECT team_id, hackathon_id, name, description, is_joinable, updated_at
FROM matchmaking.teams
WHERE team_id = $1;

-- name: ListTeamsByHackathon :many
SELECT team_id, hackathon_id, name, description, is_joinable, updated_at
FROM matchmaking.teams
WHERE hackathon_id = $1
  AND is_joinable = true;

-- name: DeleteTeam :exec
DELETE FROM matchmaking.teams
WHERE team_id = $1;
