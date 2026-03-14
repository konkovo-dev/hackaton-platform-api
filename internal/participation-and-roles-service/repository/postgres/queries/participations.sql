-- name: GetParticipation :one
SELECT * FROM participation_and_roles.participations
WHERE hackathon_id = $1 AND user_id = $2;

-- name: CreateParticipation :exec
INSERT INTO participation_and_roles.participations (
    hackathon_id,
    user_id,
    status,
    team_id,
    motivation_text,
    registered_at,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: UpdateParticipation :exec
UPDATE participation_and_roles.participations
SET status = $2,
    team_id = $3,
    updated_at = $4
WHERE hackathon_id = $1 AND user_id = $5;

-- name: UpdateParticipationProfile :exec
UPDATE participation_and_roles.participations
SET motivation_text = $2,
    updated_at = $3
WHERE hackathon_id = $1 AND user_id = $4;

-- name: UpdateParticipationStatus :exec
UPDATE participation_and_roles.participations
SET status = $2,
    updated_at = $3
WHERE hackathon_id = $1 AND user_id = $4;

-- name: DeleteParticipation :exec
DELETE FROM participation_and_roles.participations
WHERE hackathon_id = $1 AND user_id = $2;

-- name: GetWishedRolesByParticipation :many
SELECT tr.id, tr.name
FROM participation_and_roles.participation_wished_roles pwr
JOIN participation_and_roles.team_role_catalog tr ON pwr.team_role_id = tr.id
WHERE pwr.hackathon_id = $1 AND pwr.user_id = $2;

-- name: DeleteWishedRolesByParticipation :exec
DELETE FROM participation_and_roles.participation_wished_roles
WHERE hackathon_id = $1 AND user_id = $2;

-- name: CreateWishedRole :exec
INSERT INTO participation_and_roles.participation_wished_roles (hackathon_id, user_id, team_role_id)
VALUES ($1, $2, $3);

-- name: ListAllTeamRoles :many
SELECT id, name
FROM participation_and_roles.team_role_catalog
ORDER BY name ASC;

-- name: GetTeamRolesByIDs :many
SELECT id, name
FROM participation_and_roles.team_role_catalog
WHERE id = ANY($1::uuid[]);

-- name: ListParticipations :many
SELECT * FROM participation_and_roles.participations
WHERE hackathon_id = $1
ORDER BY registered_at DESC
LIMIT $2 OFFSET $3;

-- name: ListParticipationsByStatus :many
SELECT * FROM participation_and_roles.participations
WHERE hackathon_id = $1 AND status = ANY($2::TEXT[])
ORDER BY registered_at DESC
LIMIT $3 OFFSET $4;

-- name: CountParticipations :one
SELECT COUNT(*) FROM participation_and_roles.participations
WHERE hackathon_id = $1;

-- name: CountParticipationsByStatus :one
SELECT COUNT(*) FROM participation_and_roles.participations
WHERE hackathon_id = $1 AND status = ANY($2::TEXT[]);

-- name: GetHackathonIDsByUserParticipation :many
SELECT DISTINCT hackathon_id
FROM participation_and_roles.participations
WHERE user_id = $1 AND status != 'none'
ORDER BY hackathon_id;

-- name: GetHackathonIDsByUserParticipationStatus :many
SELECT DISTINCT hackathon_id
FROM participation_and_roles.participations
WHERE user_id = $1 AND status = $2
ORDER BY hackathon_id;

