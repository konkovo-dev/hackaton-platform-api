-- +goose Up
-- +goose StatementBegin

CREATE TABLE team.outbox_events (
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
    ON team.outbox_events(status, created_at)
    WHERE status = 'pending';

CREATE INDEX idx_outbox_events_aggregate 
    ON team.outbox_events(aggregate_type, aggregate_id);

CREATE INDEX idx_outbox_events_event_type 
    ON team.outbox_events(event_type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS team.outbox_events;

-- +goose StatementEnd
