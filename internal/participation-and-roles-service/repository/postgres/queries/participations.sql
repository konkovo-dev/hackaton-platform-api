-- name: GetParticipation :one
SELECT * FROM participation_and_roles.participations
WHERE hackathon_id = $1 AND user_id = $2;

-- name: CreateParticipation :exec
INSERT INTO participation_and_roles.participations (
    hackathon_id,
    user_id,
    status,
    team_id,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6
);

-- name: UpdateParticipation :exec
UPDATE participation_and_roles.participations
SET status = $2,
    team_id = $3,
    updated_at = $4
WHERE hackathon_id = $1 AND user_id = $5;

-- name: DeleteParticipation :exec
DELETE FROM participation_and_roles.participations
WHERE hackathon_id = $1 AND user_id = $2;

