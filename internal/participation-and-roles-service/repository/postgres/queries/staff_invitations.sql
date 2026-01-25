-- name: CreateStaffInvitation :exec
INSERT INTO participation_and_roles.staff_invitations (
    id,
    hackathon_id,
    target_user_id,
    requested_role,
    created_by_user_id,
    message,
    status,
    created_at,
    updated_at,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);

-- name: GetStaffInvitationByID :one
SELECT * FROM participation_and_roles.staff_invitations
WHERE id = $1;

-- name: GetPendingInvitationForUser :one
SELECT * FROM participation_and_roles.staff_invitations
WHERE hackathon_id = $1
  AND target_user_id = $2
  AND requested_role = $3
  AND status = 'pending'
LIMIT 1;

-- name: GetStaffInvitationsByTargetUser :many
SELECT * FROM participation_and_roles.staff_invitations
WHERE target_user_id = $1
ORDER BY created_at DESC;

-- name: UpdateStaffInvitationStatus :exec
UPDATE participation_and_roles.staff_invitations
SET status = $2,
    updated_at = $3
WHERE id = $1;

-- name: GetStaffInvitationsByHackathon :many
SELECT * FROM participation_and_roles.staff_invitations
WHERE hackathon_id = $1
ORDER BY created_at DESC;

