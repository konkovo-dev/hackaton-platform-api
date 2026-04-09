-- name: InsertUserStubIfNotExists :exec
INSERT INTO matchmaking.users (
    user_id,
    username,
    avatar_url,
    catalog_skill_ids,
    custom_skill_names,
    updated_at
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id) DO NOTHING;

-- name: UpsertUser :exec
INSERT INTO matchmaking.users (
    user_id,
    username,
    avatar_url,
    catalog_skill_ids,
    custom_skill_names,
    updated_at
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id) DO UPDATE SET
    username = EXCLUDED.username,
    avatar_url = EXCLUDED.avatar_url,
    catalog_skill_ids = EXCLUDED.catalog_skill_ids,
    custom_skill_names = EXCLUDED.custom_skill_names,
    updated_at = EXCLUDED.updated_at;

-- name: GetUserByID :one
SELECT user_id, username, avatar_url, catalog_skill_ids, custom_skill_names, updated_at
FROM matchmaking.users
WHERE user_id = $1;

-- name: GetUsersByIDs :many
SELECT user_id, username, avatar_url, catalog_skill_ids, custom_skill_names, updated_at
FROM matchmaking.users
WHERE user_id = ANY($1::uuid[]);
