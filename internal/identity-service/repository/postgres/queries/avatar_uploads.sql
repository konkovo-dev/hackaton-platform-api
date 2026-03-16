-- name: CreateAvatarUpload :one
INSERT INTO identity.avatar_uploads (
    upload_id,
    user_id,
    filename,
    size_bytes,
    content_type,
    storage_key,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, 'pending'
) RETURNING *;

-- name: GetAvatarUploadByID :one
SELECT * FROM identity.avatar_uploads
WHERE upload_id = $1;

-- name: CompleteAvatarUpload :one
UPDATE identity.avatar_uploads
SET status = 'completed',
    completed_at = NOW()
WHERE upload_id = $1
RETURNING *;

-- name: DeleteOldPendingAvatarUploads :exec
DELETE FROM identity.avatar_uploads
WHERE status = 'pending'
  AND created_at < NOW() - INTERVAL '24 hours';
