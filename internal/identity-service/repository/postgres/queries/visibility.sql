-- name: VisibilityGetByUserID :one
SELECT user_id, skills_visibility, contacts_visibility
FROM identity.user_visibility
WHERE user_id = $1;

-- name: VisibilityCreate :exec
INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES ($1, $2, $3);

-- name: VisibilityUpsert :exec
INSERT INTO identity.user_visibility (user_id, skills_visibility, contacts_visibility)
VALUES ($1, $2, $3)
ON CONFLICT (user_id) DO UPDATE
SET skills_visibility = EXCLUDED.skills_visibility,
    contacts_visibility = EXCLUDED.contacts_visibility;

