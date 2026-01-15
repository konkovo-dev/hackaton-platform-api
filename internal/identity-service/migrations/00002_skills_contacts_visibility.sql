-- +goose Up
-- +goose StatementBegin

CREATE TABLE identity.skill_catalog (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE identity.user_catalog_skills (
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    catalog_skill_id UUID NOT NULL REFERENCES identity.skill_catalog(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, catalog_skill_id)
);

CREATE INDEX idx_user_catalog_skills_user_id ON identity.user_catalog_skills(user_id);

CREATE TABLE identity.user_custom_skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL
);

CREATE INDEX idx_user_custom_skills_user_id ON identity.user_custom_skills(user_id);

CREATE TABLE identity.user_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    value TEXT NOT NULL,
    visibility VARCHAR(50) NOT NULL
);

CREATE INDEX idx_user_contacts_user_id ON identity.user_contacts(user_id);

CREATE TABLE identity.user_visibility (
    user_id UUID PRIMARY KEY REFERENCES identity.users(id) ON DELETE CASCADE,
    skills_visibility VARCHAR(50) NOT NULL,
    contacts_visibility VARCHAR(50) NOT NULL
);

INSERT INTO identity.skill_catalog (name) VALUES
    ('Go'),
    ('Python'),
    ('JavaScript'),
    ('TypeScript'),
    ('React'),
    ('Vue.js'),
    ('Angular'),
    ('Node.js'),
    ('PostgreSQL'),
    ('MongoDB'),
    ('Redis'),
    ('Docker'),
    ('Kubernetes'),
    ('AWS'),
    ('Git'),
    ('CI/CD'),
    ('Machine Learning'),
    ('Data Science'),
    ('UI/UX Design'),
    ('Product Management');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS identity.user_visibility;
DROP TABLE IF EXISTS identity.user_contacts;
DROP TABLE IF EXISTS identity.user_custom_skills;
DROP TABLE IF EXISTS identity.user_catalog_skills;
DROP TABLE IF EXISTS identity.skill_catalog;

-- +goose StatementEnd

