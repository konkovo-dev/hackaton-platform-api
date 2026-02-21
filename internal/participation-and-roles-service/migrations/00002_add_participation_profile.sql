-- +goose Up
-- +goose StatementBegin

CREATE TABLE participation_and_roles.team_role_catalog (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE
);

INSERT INTO participation_and_roles.team_role_catalog (id, name) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Any'),
    ('00000000-0000-0000-0000-000000000002', 'Backend'),
    ('00000000-0000-0000-0000-000000000003', 'Frontend'),
    ('00000000-0000-0000-0000-000000000004', 'Fullstack'),
    ('00000000-0000-0000-0000-000000000005', 'Mobile'),
    ('00000000-0000-0000-0000-000000000006', 'Designer'),
    ('00000000-0000-0000-0000-000000000007', 'Product Manager'),
    ('00000000-0000-0000-0000-000000000008', 'Data Scientist'),
    ('00000000-0000-0000-0000-000000000009', 'DevOps'),
    ('00000000-0000-0000-0000-00000000000a', 'QA');

ALTER TABLE participation_and_roles.participations
ADD COLUMN motivation_text TEXT NOT NULL DEFAULT '',
ADD COLUMN registered_at TIMESTAMPTZ;

UPDATE participation_and_roles.participations
SET registered_at = created_at
WHERE registered_at IS NULL;

ALTER TABLE participation_and_roles.participations
ALTER COLUMN registered_at SET NOT NULL;

CREATE TABLE participation_and_roles.participation_wished_roles (
    hackathon_id UUID NOT NULL,
    user_id UUID NOT NULL,
    team_role_id UUID NOT NULL REFERENCES participation_and_roles.team_role_catalog(id) ON DELETE CASCADE,
    PRIMARY KEY (hackathon_id, user_id, team_role_id),
    FOREIGN KEY (hackathon_id, user_id) REFERENCES participation_and_roles.participations(hackathon_id, user_id) ON DELETE CASCADE
);

CREATE INDEX idx_participation_wished_roles_participation ON participation_and_roles.participation_wished_roles(hackathon_id, user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS participation_and_roles.participation_wished_roles;

ALTER TABLE participation_and_roles.participations
DROP COLUMN IF EXISTS motivation_text,
DROP COLUMN IF EXISTS registered_at;

DROP TABLE IF EXISTS participation_and_roles.team_role_catalog;

-- +goose StatementEnd
