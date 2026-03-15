-- name: CreateEvaluation :one
INSERT INTO judging.evaluations (
    id,
    hackathon_id,
    submission_id,
    judge_user_id,
    score,
    comment,
    evaluated_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, NOW(), NOW()
)
RETURNING
    id,
    hackathon_id,
    submission_id,
    judge_user_id,
    score,
    comment,
    evaluated_at,
    updated_at;

-- name: UpdateEvaluation :one
UPDATE judging.evaluations
SET
    score = $2,
    comment = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    hackathon_id,
    submission_id,
    judge_user_id,
    score,
    comment,
    evaluated_at,
    updated_at;

-- name: GetEvaluationByID :one
SELECT
    id,
    hackathon_id,
    submission_id,
    judge_user_id,
    score,
    comment,
    evaluated_at,
    updated_at
FROM judging.evaluations
WHERE id = $1;

-- name: GetEvaluationBySubmissionAndJudge :one
SELECT
    id,
    hackathon_id,
    submission_id,
    judge_user_id,
    score,
    comment,
    evaluated_at,
    updated_at
FROM judging.evaluations
WHERE submission_id = $1 AND judge_user_id = $2;

-- name: ListEvaluationsByJudge :many
SELECT
    id,
    hackathon_id,
    submission_id,
    judge_user_id,
    score,
    comment,
    evaluated_at,
    updated_at
FROM judging.evaluations
WHERE hackathon_id = $1 AND judge_user_id = $2
ORDER BY evaluated_at DESC
LIMIT $3 OFFSET $4;

-- name: CountEvaluationsByJudge :one
SELECT COUNT(*)
FROM judging.evaluations
WHERE hackathon_id = $1 AND judge_user_id = $2;

-- name: ListEvaluationsBySubmission :many
SELECT
    id,
    hackathon_id,
    submission_id,
    judge_user_id,
    score,
    comment,
    evaluated_at,
    updated_at
FROM judging.evaluations
WHERE submission_id = $1
ORDER BY evaluated_at ASC;

-- name: GetLeaderboardScores :many
SELECT
    e.submission_id,
    COALESCE(AVG(e.score::FLOAT), 0) as average_score,
    COUNT(e.id)::INT as evaluation_count,
    MIN(e.evaluated_at) as first_evaluated_at
FROM judging.evaluations e
WHERE e.hackathon_id = $1
GROUP BY e.submission_id
ORDER BY average_score DESC NULLS LAST, first_evaluated_at ASC
LIMIT $2 OFFSET $3;

-- name: CountEvaluatedSubmissions :one
SELECT COUNT(DISTINCT e.submission_id)
FROM judging.evaluations e
WHERE e.hackathon_id = $1;

-- name: GetSubmissionAverageScore :one
SELECT
    COALESCE(AVG(e.score::FLOAT), 0) as average_score,
    COUNT(e.id)::INT as evaluation_count
FROM judging.evaluations e
WHERE e.submission_id = $1
GROUP BY e.submission_id;

-- name: GetEvaluationsByOwner :many
SELECT
    e.id,
    e.hackathon_id,
    e.submission_id,
    e.judge_user_id,
    e.score,
    e.comment,
    e.evaluated_at,
    e.updated_at
FROM judging.evaluations e
WHERE e.hackathon_id = $1 AND e.submission_id = $2
ORDER BY e.evaluated_at ASC;
