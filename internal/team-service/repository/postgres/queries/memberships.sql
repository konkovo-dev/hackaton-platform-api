-- name: CreateMembership :one
INSERT INTO team.memberships (
    team_id,
    user_id,
    is_captain,
    assigned_vacancy_id,
    joined_at
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING
    team_id,
    user_id,
    is_captain,
    assigned_vacancy_id,
    joined_at;

-- name: CheckIsCaptain :one
SELECT EXISTS(
    SELECT 1
    FROM team.memberships
    WHERE team_id = $1 AND user_id = $2 AND is_captain = true
);

-- name: CountMembers :one
SELECT COUNT(*)
FROM team.memberships
WHERE team_id = $1;

-- name: ListTeamMembers :many
SELECT
    team_id,
    user_id,
    is_captain,
    assigned_vacancy_id,
    joined_at
FROM team.memberships
WHERE team_id = $1
ORDER BY is_captain DESC, joined_at ASC;

-- name: GetMembership :one
SELECT
    team_id,
    user_id,
    is_captain,
    assigned_vacancy_id,
    joined_at
FROM team.memberships
WHERE team_id = $1 AND user_id = $2;

-- name: DeleteMembership :exec
DELETE FROM team.memberships
WHERE team_id = $1 AND user_id = $2;

-- name: UpdateCaptainStatus :exec
UPDATE team.memberships
SET is_captain = $3
WHERE team_id = $1 AND user_id = $2;
