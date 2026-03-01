-- name: UpsertParticipation :exec
INSERT INTO matchmaking.participations (
    hackathon_id,
    user_id,
    status,
    wished_role_ids,
    motivation_text,
    team_id,
    updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (hackathon_id, user_id) DO UPDATE SET
    status = EXCLUDED.status,
    wished_role_ids = EXCLUDED.wished_role_ids,
    motivation_text = EXCLUDED.motivation_text,
    team_id = EXCLUDED.team_id,
    updated_at = EXCLUDED.updated_at;

-- name: GetParticipation :one
SELECT hackathon_id, user_id, status, wished_role_ids, motivation_text, team_id, updated_at
FROM matchmaking.participations
WHERE hackathon_id = $1 AND user_id = $2;

-- name: ListParticipationsByHackathon :many
SELECT hackathon_id, user_id, status, wished_role_ids, motivation_text, team_id, updated_at
FROM matchmaking.participations
WHERE hackathon_id = $1
  AND team_id IS NULL
  AND status IN ('looking_for_team', 'individual');

-- name: DeleteParticipation :exec
DELETE FROM matchmaking.participations
WHERE hackathon_id = $1 AND user_id = $2;
