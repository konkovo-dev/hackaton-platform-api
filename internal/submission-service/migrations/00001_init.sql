-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS submission;

CREATE TABLE submission.submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hackathon_id UUID NOT NULL,
    owner_kind VARCHAR(20) NOT NULL CHECK (owner_kind IN ('user', 'team')),
    owner_id UUID NOT NULL,
    created_by_user_id UUID NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_final BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_submissions_one_final_per_owner 
    ON submission.submissions(hackathon_id, owner_kind, owner_id) 
    WHERE is_final = true;

CREATE INDEX idx_submissions_hackathon_owner 
    ON submission.submissions(hackathon_id, owner_kind, owner_id, created_at DESC);

CREATE INDEX idx_submissions_created_by 
    ON submission.submissions(created_by_user_id);

CREATE TABLE submission.submission_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id UUID NOT NULL REFERENCES submission.submissions(id) ON DELETE CASCADE,
    filename VARCHAR(500) NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes > 0),
    content_type VARCHAR(100) NOT NULL,
    storage_key TEXT NOT NULL,
    upload_status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (upload_status IN ('pending', 'completed', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_submission_files_submission_id 
    ON submission.submission_files(submission_id);

CREATE INDEX idx_submission_files_status 
    ON submission.submission_files(upload_status);

CREATE INDEX idx_submission_files_pending_cleanup 
    ON submission.submission_files(created_at) 
    WHERE upload_status = 'pending';

CREATE TABLE submission.idempotency_keys (
    key TEXT NOT NULL,
    scope TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    response_blob BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (key, scope)
);

CREATE INDEX idx_idempotency_keys_expires_at 
    ON submission.idempotency_keys(expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS submission.idempotency_keys;
DROP TABLE IF EXISTS submission.submission_files;
DROP TABLE IF EXISTS submission.submissions;
DROP SCHEMA IF EXISTS submission;

-- +goose StatementEnd
