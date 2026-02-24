-- +goose Up
-- +goose StatementBegin

DROP INDEX IF EXISTS mentors.tickets_hackathon_id_owner_kind_owner_id_status_key;

CREATE UNIQUE INDEX idx_one_open_ticket_per_owner 
ON mentors.tickets(hackathon_id, owner_kind, owner_id) 
WHERE status = 'open';

DROP INDEX IF EXISTS mentors.messages_client_message_id_key;

CREATE UNIQUE INDEX idx_unique_client_message_per_ticket 
ON mentors.messages(ticket_id, client_message_id) 
WHERE client_message_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS mentors.idx_one_open_ticket_per_owner;
DROP INDEX IF EXISTS mentors.idx_unique_client_message_per_ticket;

CREATE UNIQUE INDEX tickets_hackathon_id_owner_kind_owner_id_status_key 
ON mentors.tickets(hackathon_id, owner_kind, owner_id, status);

CREATE UNIQUE INDEX messages_client_message_id_key 
ON mentors.messages(client_message_id);

-- +goose StatementEnd
