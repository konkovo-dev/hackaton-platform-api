-- +goose Up
ALTER TABLE mentors.messages ALTER COLUMN author_user_id DROP NOT NULL;
ALTER TABLE mentors.messages DROP CONSTRAINT messages_author_role_check;
ALTER TABLE mentors.messages ADD CONSTRAINT messages_author_role_check 
    CHECK (author_role IN ('participant', 'mentor', 'organizer', 'system'));

-- +goose Down
DELETE FROM mentors.messages WHERE author_role = 'system';
ALTER TABLE mentors.messages DROP CONSTRAINT messages_author_role_check;
ALTER TABLE mentors.messages ADD CONSTRAINT messages_author_role_check 
    CHECK (author_role IN ('participant', 'mentor', 'organizer'));
ALTER TABLE mentors.messages ALTER COLUMN author_user_id SET NOT NULL;
