-- name: ContactGetByUserID :many
SELECT id, user_id, type, value, visibility
FROM identity.user_contacts
WHERE user_id = $1;

-- name: ContactDeleteByUserID :exec
DELETE FROM identity.user_contacts
WHERE user_id = $1;

-- name: ContactCreate :exec
INSERT INTO identity.user_contacts (id, user_id, type, value, visibility)
VALUES ($1, $2, $3, $4, $5);

-- name: ContactGetByUserIDs :many
SELECT id, user_id, type, value, visibility
FROM identity.user_contacts
WHERE user_id = ANY($1::uuid[]);
