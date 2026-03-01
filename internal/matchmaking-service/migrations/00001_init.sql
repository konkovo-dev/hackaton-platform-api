-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA matchmaking;

CREATE TABLE matchmaking.users (
    user_id UUID PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    catalog_skill_ids UUID[] NOT NULL DEFAULT '{}',
    custom_skill_names TEXT[] NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE matchmaking.participations (
    hackathon_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    wished_role_ids UUID[] NOT NULL DEFAULT '{}',
    motivation_text TEXT,
    motivation_tsv tsvector GENERATED ALWAYS AS (
        to_tsvector('english', COALESCE(motivation_text, '')) || 
        to_tsvector('russian', COALESCE(motivation_text, ''))
    ) STORED,
    team_id UUID,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (hackathon_id, user_id)
);

CREATE TABLE matchmaking.teams (
    team_id UUID PRIMARY KEY,
    hackathon_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    description_tsv tsvector GENERATED ALWAYS AS (
        to_tsvector('english', COALESCE(description, '')) || 
        to_tsvector('russian', COALESCE(description, ''))
    ) STORED,
    is_joinable BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE matchmaking.vacancies (
    vacancy_id UUID PRIMARY KEY,
    team_id UUID NOT NULL REFERENCES matchmaking.teams(team_id) ON DELETE CASCADE,
    hackathon_id UUID NOT NULL,
    description TEXT,
    description_tsv tsvector GENERATED ALWAYS AS (
        to_tsvector('english', COALESCE(description, '')) || 
        to_tsvector('russian', COALESCE(description, ''))
    ) STORED,
    desired_role_ids UUID[] NOT NULL DEFAULT '{}',
    desired_skill_ids UUID[] NOT NULL DEFAULT '{}',
    slots_open INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_participations_motivation_tsv ON matchmaking.participations USING GIN(motivation_tsv);
CREATE INDEX idx_teams_description_tsv ON matchmaking.teams USING GIN(description_tsv);
CREATE INDEX idx_vacancies_description_tsv ON matchmaking.vacancies USING GIN(description_tsv);

CREATE INDEX idx_participations_hackathon_status ON matchmaking.participations(hackathon_id, status) WHERE team_id IS NULL;
CREATE INDEX idx_teams_hackathon_joinable ON matchmaking.teams(hackathon_id) WHERE is_joinable = true;
CREATE INDEX idx_vacancies_hackathon_open ON matchmaking.vacancies(hackathon_id) WHERE slots_open > 0;

CREATE TABLE matchmaking.outbox_events (
    id UUID PRIMARY KEY,
    aggregate_id TEXT NOT NULL,
    aggregate_type TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    attempt_count INT NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outbox_events_status_created 
    ON matchmaking.outbox_events(status, created_at)
    WHERE status = 'pending';

CREATE INDEX idx_outbox_events_aggregate 
    ON matchmaking.outbox_events(aggregate_type, aggregate_id);

CREATE INDEX idx_outbox_events_event_type 
    ON matchmaking.outbox_events(event_type);

CREATE TABLE matchmaking.idempotency_keys (
    key TEXT NOT NULL,
    scope TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    response_blob BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (key, scope)
);

CREATE INDEX idx_idempotency_keys_expires_at ON matchmaking.idempotency_keys(expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS matchmaking.idempotency_keys;
DROP TABLE IF EXISTS matchmaking.outbox_events;
DROP TABLE IF EXISTS matchmaking.vacancies;
DROP TABLE IF EXISTS matchmaking.teams;
DROP TABLE IF EXISTS matchmaking.participations;
DROP TABLE IF EXISTS matchmaking.users;
DROP SCHEMA IF EXISTS matchmaking;

-- +goose StatementEnd
