-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS mentors;

CREATE TABLE mentors.tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hackathon_id UUID NOT NULL,
    owner_kind VARCHAR(20) NOT NULL CHECK (owner_kind IN ('user', 'team')),
    owner_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed')),
    assigned_mentor_user_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_tickets_one_open_per_owner 
    ON mentors.tickets(hackathon_id, owner_kind, owner_id) 
    WHERE status = 'open';

CREATE INDEX idx_tickets_hackathon_id ON mentors.tickets(hackathon_id);
CREATE INDEX idx_tickets_assigned_mentor ON mentors.tickets(assigned_mentor_user_id) WHERE assigned_mentor_user_id IS NOT NULL;
CREATE INDEX idx_tickets_status ON mentors.tickets(status);
CREATE INDEX idx_tickets_created_at ON mentors.tickets(created_at DESC);
CREATE INDEX idx_tickets_updated_at ON mentors.tickets(updated_at DESC);

CREATE TABLE mentors.messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES mentors.tickets(id) ON DELETE CASCADE,
    author_user_id UUID NOT NULL,
    author_role VARCHAR(20) NOT NULL CHECK (author_role IN ('participant', 'mentor', 'organizer')),
    text TEXT NOT NULL,
    client_message_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_ticket_id ON mentors.messages(ticket_id);
CREATE INDEX idx_messages_created_at ON mentors.messages(created_at ASC);
CREATE INDEX idx_messages_client_id ON mentors.messages(client_message_id) WHERE client_message_id IS NOT NULL;

CREATE TABLE mentors.idempotency_keys (
    key TEXT NOT NULL,
    scope TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    response_blob BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (key, scope)
);

CREATE INDEX idx_idempotency_keys_expires_at ON mentors.idempotency_keys(expires_at);

CREATE TABLE mentors.outbox_events (
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
    ON mentors.outbox_events(status, created_at)
    WHERE status = 'pending';

CREATE INDEX idx_outbox_events_aggregate 
    ON mentors.outbox_events(aggregate_type, aggregate_id);

CREATE INDEX idx_outbox_events_event_type 
    ON mentors.outbox_events(event_type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS mentors.outbox_events;
DROP TABLE IF EXISTS mentors.idempotency_keys;
DROP TABLE IF EXISTS mentors.messages;
DROP TABLE IF EXISTS mentors.tickets;
DROP SCHEMA IF EXISTS mentors;

-- +goose StatementEnd
