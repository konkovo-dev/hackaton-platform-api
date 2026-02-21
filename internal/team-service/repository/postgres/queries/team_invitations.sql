-- name: ListTeamInvitations :many
SELECT
    id,
    hackathon_id,
    team_id,
    vacancy_id,
    target_user_id,
    created_by_user_id,
    message,
    status,
    created_at,
    updated_at,
    expires_at
FROM team.team_invitations
WHERE team_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTeamInvitations :one
SELECT COUNT(*)
FROM team.team_invitations
WHERE team_id = $1;

-- name: CreateTeamInvitation :one
INSERT INTO team.team_invitations (
    id,
    hackathon_id,
    team_id,
    vacancy_id,
    target_user_id,
    created_by_user_id,
    message,
    status,
    created_at,
    updated_at,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING
    id,
    hackathon_id,
    team_id,
    vacancy_id,
    target_user_id,
    created_by_user_id,
    message,
    status,
    created_at,
    updated_at,
    expires_at;

-- name: CheckUserInTeam :one
SELECT EXISTS(
    SELECT 1
    FROM team.memberships
    WHERE team_id = $1 AND user_id = $2
);

-- name: GetTeamInvitationByID :one
SELECT
    id,
    hackathon_id,
    team_id,
    vacancy_id,
    target_user_id,
    created_by_user_id,
    message,
    status,
    created_at,
    updated_at,
    expires_at
FROM team.team_invitations
WHERE id = $1;

-- name: UpdateTeamInvitationStatus :one
UPDATE team.team_invitations
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    hackathon_id,
    team_id,
    vacancy_id,
    target_user_id,
    created_by_user_id,
    message,
    status,
    created_at,
    updated_at,
    expires_at;

-- name: ListMyTeamInvitations :many
SELECT
    id,
    hackathon_id,
    team_id,
    vacancy_id,
    target_user_id,
    created_by_user_id,
    message,
    status,
    created_at,
    updated_at,
    expires_at
FROM team.team_invitations
WHERE target_user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountMyTeamInvitations :one
SELECT COUNT(*)
FROM team.team_invitations
WHERE target_user_id = $1;

-- name: CancelCompetingInvitations :exec
UPDATE team.team_invitations
SET
    status = 'canceled',
    updated_at = NOW()
WHERE target_user_id = $1
  AND hackathon_id = $2
  AND status = 'pending';
