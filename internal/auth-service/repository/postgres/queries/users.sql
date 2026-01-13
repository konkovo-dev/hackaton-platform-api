-- name: CreateUser :exec
INSERT INTO auth.users (id, username, email, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUserByID :one
SELECT id, username, email, created_at, updated_at
FROM auth.users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, email, created_at, updated_at
FROM auth.users
WHERE username = $1;

-- name: GetUserByEmail :one
SELECT id, username, email, created_at, updated_at
FROM auth.users
WHERE email = $1;

-- name: UpdateUser :exec
UPDATE auth.users
SET email = $2, updated_at = $3
WHERE id = $1;

