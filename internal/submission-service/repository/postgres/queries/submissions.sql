-- name: CreateSubmission :one
INSERT INTO submission.submissions (
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    created_by_user_id,
    title,
    description,
    is_final,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
)
RETURNING
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    created_by_user_id,
    title,
    description,
    is_final,
    created_at,
    updated_at;

-- name: GetSubmissionByID :one
SELECT
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    created_by_user_id,
    title,
    description,
    is_final,
    created_at,
    updated_at
FROM submission.submissions
WHERE id = $1;

-- name: GetSubmissionByIDAndHackathonID :one
SELECT
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    created_by_user_id,
    title,
    description,
    is_final,
    created_at,
    updated_at
FROM submission.submissions
WHERE id = $1 AND hackathon_id = $2;

-- name: ListSubmissionsByOwner :many
SELECT
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    created_by_user_id,
    title,
    description,
    is_final,
    created_at,
    updated_at
FROM submission.submissions
WHERE hackathon_id = $1 AND owner_kind = $2 AND owner_id = $3
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListSubmissionsByHackathon :many
SELECT
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    created_by_user_id,
    title,
    description,
    is_final,
    created_at,
    updated_at
FROM submission.submissions
WHERE hackathon_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountSubmissionsByOwner :one
SELECT COUNT(*)
FROM submission.submissions
WHERE hackathon_id = $1 AND owner_kind = $2 AND owner_id = $3;

-- name: GetFinalSubmissionByOwner :one
SELECT
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    created_by_user_id,
    title,
    description,
    is_final,
    created_at,
    updated_at
FROM submission.submissions
WHERE hackathon_id = $1 AND owner_kind = $2 AND owner_id = $3 AND is_final = true
LIMIT 1;

-- name: UpdateSubmissionDescription :one
UPDATE submission.submissions
SET
    description = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    hackathon_id,
    owner_kind,
    owner_id,
    created_by_user_id,
    title,
    description,
    is_final,
    created_at,
    updated_at;

-- name: UnsetFinalSubmission :exec
UPDATE submission.submissions
SET
    is_final = false,
    updated_at = NOW()
WHERE hackathon_id = $1 AND owner_kind = $2 AND owner_id = $3 AND is_final = true;

-- name: SetFinalSubmission :exec
UPDATE submission.submissions
SET
    is_final = true,
    updated_at = NOW()
WHERE id = $1;

-- name: GetSubmissionTotalSize :one
SELECT COALESCE(SUM(size_bytes), 0)::BIGINT
FROM submission.submission_files
WHERE submission_id = $1 AND upload_status = 'completed';
