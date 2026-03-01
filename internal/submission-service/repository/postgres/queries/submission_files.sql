-- name: CreateSubmissionFile :one
INSERT INTO submission.submission_files (
    id,
    submission_id,
    filename,
    size_bytes,
    content_type,
    storage_key,
    upload_status,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, NOW()
)
RETURNING
    id,
    submission_id,
    filename,
    size_bytes,
    content_type,
    storage_key,
    upload_status,
    created_at,
    completed_at;

-- name: GetFileByID :one
SELECT
    id,
    submission_id,
    filename,
    size_bytes,
    content_type,
    storage_key,
    upload_status,
    created_at,
    completed_at
FROM submission.submission_files
WHERE id = $1;

-- name: GetFileByIDAndSubmissionID :one
SELECT
    id,
    submission_id,
    filename,
    size_bytes,
    content_type,
    storage_key,
    upload_status,
    created_at,
    completed_at
FROM submission.submission_files
WHERE id = $1 AND submission_id = $2;

-- name: ListFilesBySubmission :many
SELECT
    id,
    submission_id,
    filename,
    size_bytes,
    content_type,
    storage_key,
    upload_status,
    created_at,
    completed_at
FROM submission.submission_files
WHERE submission_id = $1
ORDER BY created_at ASC;

-- name: CountFilesBySubmission :one
SELECT COUNT(*)
FROM submission.submission_files
WHERE submission_id = $1;

-- name: CountCompletedFilesBySubmission :one
SELECT COUNT(*)
FROM submission.submission_files
WHERE submission_id = $1 AND upload_status = 'completed';

-- name: UpdateFileStatus :one
UPDATE submission.submission_files
SET
    upload_status = $2,
    completed_at = $3
WHERE id = $1
RETURNING
    id,
    submission_id,
    filename,
    size_bytes,
    content_type,
    storage_key,
    upload_status,
    created_at,
    completed_at;

-- name: MarkExpiredFilesAsFailed :execrows
UPDATE submission.submission_files
SET
    upload_status = 'failed',
    completed_at = NOW()
WHERE upload_status = 'pending'
  AND created_at < NOW() - $1::INTERVAL;
