-- name: UserCreate :exec
INSERT INTO identity.users (id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: UserGetByID :one
SELECT id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at
FROM identity.users
WHERE id = $1;

-- name: UserGetByUsername :one
SELECT id, username, first_name, last_name, avatar_url, timezone, created_at, updated_at
FROM identity.users
WHERE username = $1;

-- name: UserUpdate :exec
UPDATE identity.users
SET first_name = $2,
    last_name = $3,
    avatar_url = $4,
    timezone = $5,
    updated_at = $6
WHERE id = $1;

