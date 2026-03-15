-- name: CreateAssignment :one
INSERT INTO judging.assignments (
    id,
    hackathon_id,
    submission_id,
    judge_user_id,
    assigned_at
) VALUES (
    $1, $2, $3, $4, NOW()
)
RETURNING
    id,
    hackathon_id,
    submission_id,
    judge_user_id,
    assigned_at;

-- name: GetAssignmentByID :one
SELECT
    id,
    hackathon_id,
    submission_id,
    judge_user_id,
    assigned_at
FROM judging.assignments
WHERE id = $1;

-- name: ListAssignmentsByJudge :many
SELECT
    a.id,
    a.hackathon_id,
    a.submission_id,
    a.judge_user_id,
    a.assigned_at,
    CASE WHEN e.id IS NOT NULL THEN true ELSE false END as is_evaluated
FROM judging.assignments a
LEFT JOIN judging.evaluations e ON e.submission_id = a.submission_id AND e.judge_user_id = a.judge_user_id
WHERE a.hackathon_id = $1 AND a.judge_user_id = $2
ORDER BY a.assigned_at ASC
LIMIT $3 OFFSET $4;

-- name: ListAssignmentsByJudgeFiltered :many
SELECT
    a.id,
    a.hackathon_id,
    a.submission_id,
    a.judge_user_id,
    a.assigned_at,
    CASE WHEN e.id IS NOT NULL THEN true ELSE false END as is_evaluated
FROM judging.assignments a
LEFT JOIN judging.evaluations e ON e.submission_id = a.submission_id AND e.judge_user_id = a.judge_user_id
WHERE a.hackathon_id = $1 AND a.judge_user_id = $2 
  AND (CASE WHEN e.id IS NOT NULL THEN true ELSE false END)::boolean = sqlc.arg(evaluated)::boolean
ORDER BY a.assigned_at ASC
LIMIT $3 OFFSET $4;

-- name: CountAssignmentsByJudge :one
SELECT COUNT(*)
FROM judging.assignments
WHERE hackathon_id = $1 AND judge_user_id = $2;

-- name: CountAssignmentsByJudgeFiltered :one
SELECT COUNT(*)
FROM judging.assignments a
LEFT JOIN judging.evaluations e ON e.submission_id = a.submission_id AND e.judge_user_id = a.judge_user_id
WHERE a.hackathon_id = $1 AND a.judge_user_id = $2 
  AND (CASE WHEN e.id IS NOT NULL THEN true ELSE false END)::boolean = sqlc.arg(evaluated)::boolean;

-- name: CheckAssignmentExists :one
SELECT EXISTS(
    SELECT 1 FROM judging.assignments
    WHERE hackathon_id = $1 AND submission_id = $2 AND judge_user_id = $3
);

-- name: CountAssignmentsByHackathon :one
SELECT COUNT(*)
FROM judging.assignments
WHERE hackathon_id = $1;
