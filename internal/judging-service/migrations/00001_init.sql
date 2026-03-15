-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS judging;

CREATE TABLE judging.assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hackathon_id UUID NOT NULL,
    submission_id UUID NOT NULL,
    judge_user_id UUID NOT NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(hackathon_id, submission_id, judge_user_id)
);

CREATE INDEX idx_assignments_hackathon 
    ON judging.assignments(hackathon_id);

CREATE INDEX idx_assignments_judge 
    ON judging.assignments(judge_user_id, hackathon_id);

CREATE INDEX idx_assignments_submission 
    ON judging.assignments(submission_id);

CREATE TABLE judging.evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hackathon_id UUID NOT NULL,
    submission_id UUID NOT NULL,
    judge_user_id UUID NOT NULL,
    score INT NOT NULL CHECK (score >= 0 AND score <= 10),
    comment TEXT NOT NULL CHECK (LENGTH(comment) >= 1 AND LENGTH(comment) <= 5000),
    evaluated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(hackathon_id, submission_id, judge_user_id)
);

CREATE INDEX idx_evaluations_hackathon 
    ON judging.evaluations(hackathon_id);

CREATE INDEX idx_evaluations_judge 
    ON judging.evaluations(judge_user_id, hackathon_id);

CREATE INDEX idx_evaluations_submission 
    ON judging.evaluations(submission_id);

CREATE TABLE judging.idempotency_keys (
    key TEXT NOT NULL,
    scope TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    response_blob BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (key, scope)
);

CREATE INDEX idx_idempotency_keys_expires_at 
    ON judging.idempotency_keys(expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS judging.idempotency_keys;
DROP TABLE IF EXISTS judging.evaluations;
DROP TABLE IF EXISTS judging.assignments;
DROP SCHEMA IF EXISTS judging;

-- +goose StatementEnd
