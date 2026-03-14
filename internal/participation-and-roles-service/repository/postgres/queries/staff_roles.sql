-- name: CreateStaffRole :exec
INSERT INTO participation_and_roles.staff_roles (
    hackathon_id,
    user_id,
    role,
    created_at
) VALUES (
    $1, $2, $3, NOW()
) ON CONFLICT (hackathon_id, user_id, role) DO NOTHING;

-- name: GetStaffRolesByHackathonID :many
SELECT hackathon_id, user_id, role, created_at
FROM participation_and_roles.staff_roles
WHERE hackathon_id = $1
ORDER BY created_at ASC;

-- name: GetStaffRolesByUserID :many
SELECT hackathon_id, user_id, role, created_at
FROM participation_and_roles.staff_roles
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetStaffRolesByHackathonAndUser :many
SELECT hackathon_id, user_id, role, created_at
FROM participation_and_roles.staff_roles
WHERE hackathon_id = $1 AND user_id = $2;

-- name: DeleteStaffRole :exec
DELETE FROM participation_and_roles.staff_roles
WHERE hackathon_id = $1 AND user_id = $2 AND role = $3;

-- name: HasStaffRole :one
SELECT EXISTS(
    SELECT 1
    FROM participation_and_roles.staff_roles
    WHERE hackathon_id = $1 AND user_id = $2 AND role = $3
);

-- name: GetHackathonIDsByUserRole :many
SELECT DISTINCT hackathon_id
FROM participation_and_roles.staff_roles
WHERE user_id = $1 AND role = $2
ORDER BY hackathon_id;

-- name: GetHackathonIDsByUserAnyRole :many
SELECT DISTINCT hackathon_id
FROM participation_and_roles.staff_roles
WHERE user_id = $1
ORDER BY hackathon_id;

