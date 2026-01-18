-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS hackathon;

CREATE TABLE hackathon.hackathons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    short_description TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    
    location_online BOOLEAN NOT NULL DEFAULT false,
    location_city VARCHAR(255) NOT NULL DEFAULT '',
    location_country VARCHAR(255) NOT NULL DEFAULT '',
    location_venue TEXT NOT NULL DEFAULT '',
    
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    registration_opens_at TIMESTAMPTZ,
    registration_closes_at TIMESTAMPTZ,
    submissions_opens_at TIMESTAMPTZ,
    submissions_closes_at TIMESTAMPTZ,
    judging_ends_at TIMESTAMPTZ,
    
    stage VARCHAR(50) NOT NULL DEFAULT 'hackathon_stage_upcoming',
    
    team_size_max INT NOT NULL DEFAULT 0,
    
    allow_individual BOOLEAN NOT NULL DEFAULT true,
    allow_team BOOLEAN NOT NULL DEFAULT true,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hackathons_name ON hackathon.hackathons(name);
CREATE INDEX idx_hackathons_stage ON hackathon.hackathons(stage);
CREATE INDEX idx_hackathons_created_at ON hackathon.hackathons(created_at DESC);
CREATE INDEX idx_hackathons_starts_at ON hackathon.hackathons(starts_at);

CREATE TABLE hackathon.hackathon_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hackathon_id UUID NOT NULL REFERENCES hackathon.hackathons(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    url TEXT NOT NULL
);

CREATE INDEX idx_hackathon_links_hackathon_id ON hackathon.hackathon_links(hackathon_id);

CREATE TABLE hackathon.idempotency_keys (
    key TEXT NOT NULL,
    scope TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    response_blob BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (key, scope)
);

CREATE INDEX idx_idempotency_keys_expires_at ON hackathon.idempotency_keys(expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS hackathon.idempotency_keys;
DROP TABLE IF EXISTS hackathon.hackathon_links;
DROP TABLE IF EXISTS hackathon.hackathons;
DROP SCHEMA IF EXISTS hackathon;

-- +goose StatementEnd

