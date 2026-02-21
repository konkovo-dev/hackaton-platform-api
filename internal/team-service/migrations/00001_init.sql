-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS team;

CREATE TABLE team.teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hackathon_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_joinable BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (hackathon_id, name)
);

CREATE INDEX idx_teams_hackathon_id ON team.teams(hackathon_id);
CREATE INDEX idx_teams_created_at ON team.teams(created_at DESC);

CREATE TABLE team.vacancies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES team.teams(id) ON DELETE CASCADE,
    description TEXT NOT NULL DEFAULT '',
    desired_role_ids UUID[] NOT NULL DEFAULT '{}',
    desired_skill_ids UUID[] NOT NULL DEFAULT '{}',
    slots_total BIGINT NOT NULL DEFAULT 1,
    slots_open BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (slots_open >= 0),
    CHECK (slots_open <= slots_total)
);

CREATE INDEX idx_vacancies_team_id ON team.vacancies(team_id);

CREATE TABLE team.memberships (
    team_id UUID NOT NULL REFERENCES team.teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    is_captain BOOLEAN NOT NULL DEFAULT false,
    assigned_vacancy_id UUID REFERENCES team.vacancies(id) ON DELETE SET NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (team_id, user_id)
);

CREATE INDEX idx_memberships_user_id ON team.memberships(user_id);
CREATE INDEX idx_memberships_team_id ON team.memberships(team_id);

CREATE TABLE team.team_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hackathon_id UUID NOT NULL,
    team_id UUID NOT NULL REFERENCES team.teams(id) ON DELETE CASCADE,
    vacancy_id UUID NOT NULL REFERENCES team.vacancies(id) ON DELETE CASCADE,
    target_user_id UUID NOT NULL,
    created_by_user_id UUID NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'canceled', 'expired')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_team_invitations_hackathon_id ON team.team_invitations(hackathon_id);
CREATE INDEX idx_team_invitations_team_id ON team.team_invitations(team_id);
CREATE INDEX idx_team_invitations_target_user_id ON team.team_invitations(target_user_id);
CREATE INDEX idx_team_invitations_status ON team.team_invitations(status);
CREATE INDEX idx_team_invitations_created_at ON team.team_invitations(created_at DESC);

CREATE TABLE team.join_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hackathon_id UUID NOT NULL,
    team_id UUID NOT NULL REFERENCES team.teams(id) ON DELETE CASCADE,
    vacancy_id UUID NOT NULL REFERENCES team.vacancies(id) ON DELETE CASCADE,
    requester_user_id UUID NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'canceled', 'expired')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_join_requests_hackathon_id ON team.join_requests(hackathon_id);
CREATE INDEX idx_join_requests_team_id ON team.join_requests(team_id);
CREATE INDEX idx_join_requests_requester_user_id ON team.join_requests(requester_user_id);
CREATE INDEX idx_join_requests_status ON team.join_requests(status);
CREATE INDEX idx_join_requests_created_at ON team.join_requests(created_at DESC);

CREATE TABLE team.idempotency_keys (
    key TEXT NOT NULL,
    scope TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    response_blob BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (key, scope)
);

CREATE INDEX idx_idempotency_keys_expires_at ON team.idempotency_keys(expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS team.idempotency_keys;
DROP TABLE IF EXISTS team.join_requests;
DROP TABLE IF EXISTS team.team_invitations;
DROP TABLE IF EXISTS team.memberships;
DROP TABLE IF EXISTS team.vacancies;
DROP TABLE IF EXISTS team.teams;
DROP SCHEMA IF EXISTS team;

-- +goose StatementEnd
