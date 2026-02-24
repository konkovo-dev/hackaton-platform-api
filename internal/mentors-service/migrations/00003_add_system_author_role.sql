-- +goose Up
-- +goose StatementBegin

ALTER TABLE mentors.messages 
DROP CONSTRAINT IF EXISTS messages_author_role_check;

ALTER TABLE mentors.messages 
ADD CONSTRAINT messages_author_role_check 
CHECK (author_role IN ('participant', 'mentor', 'organizer', 'system'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE mentors.messages 
DROP CONSTRAINT IF EXISTS messages_author_role_check;

ALTER TABLE mentors.messages 
ADD CONSTRAINT messages_author_role_check 
CHECK (author_role IN ('participant', 'mentor', 'organizer'));

-- +goose StatementEnd
