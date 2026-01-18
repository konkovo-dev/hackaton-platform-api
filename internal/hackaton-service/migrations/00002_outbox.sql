-- +goose Up
-- +goose StatementBegin

CREATE TABLE hackathon.outbox_events (
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
    ON hackathon.outbox_events(status, created_at)
    WHERE status = 'pending';

CREATE INDEX idx_outbox_events_aggregate 
    ON hackathon.outbox_events(aggregate_type, aggregate_id);

CREATE INDEX idx_outbox_events_event_type 
    ON hackathon.outbox_events(event_type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS hackathon.idx_outbox_events_event_type;
DROP INDEX IF EXISTS hackathon.idx_outbox_events_aggregate;
DROP INDEX IF EXISTS hackathon.idx_outbox_events_status_created;
DROP TABLE IF EXISTS hackathon.outbox_events;

-- +goose StatementEnd

