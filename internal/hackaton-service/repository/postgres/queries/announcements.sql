-- name: CreateAnnouncement :exec
INSERT INTO hackathon.announcements (
    id,
    hackathon_id,
    title,
    body,
    created_by_user_id,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
);

-- name: GetAnnouncementByID :one
SELECT 
    id,
    hackathon_id,
    title,
    body,
    created_by_user_id,
    created_at,
    updated_at,
    deleted_at
FROM hackathon.announcements
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListAnnouncementsByHackathonID :many
SELECT 
    id,
    hackathon_id,
    title,
    body,
    created_by_user_id,
    created_at,
    updated_at,
    deleted_at
FROM hackathon.announcements
WHERE hackathon_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateAnnouncement :exec
UPDATE hackathon.announcements
SET 
    title = $2,
    body = $3,
    updated_at = $4
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteAnnouncement :exec
UPDATE hackathon.announcements
SET 
    deleted_at = $2,
    updated_at = $2
WHERE id = $1 AND deleted_at IS NULL;

