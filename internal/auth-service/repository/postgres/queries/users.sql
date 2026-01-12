-- name: CreateUser :exec
INSERT INTO auth.users (id, username, email, first_name, last_name, timezone, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetUserByID :one
SELECT id, username, email, first_name, last_name, timezone, created_at, updated_at
FROM auth.users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, email, first_name, last_name, timezone, created_at, updated_at
FROM auth.users
WHERE username = $1;

-- name: GetUserByEmail :one
SELECT id, username, email, first_name, last_name, timezone, created_at, updated_at
FROM auth.users
WHERE email = $1;

-- name: UpdateUser :exec
UPDATE auth.users
SET email = $2, first_name = $3, last_name = $4, timezone = $5, updated_at = $6
WHERE id = $1;

