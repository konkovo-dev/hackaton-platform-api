-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS participation_and_roles;

CREATE TABLE participation_and_roles.staff_roles (
    hackathon_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('owner', 'organizer', 'mentor', 'judge')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (hackathon_id, user_id, role)
);

CREATE INDEX idx_staff_roles_hackathon_id ON participation_and_roles.staff_roles(hackathon_id);
CREATE INDEX idx_staff_roles_user_id ON participation_and_roles.staff_roles(user_id);
CREATE INDEX idx_staff_roles_role ON participation_and_roles.staff_roles(role);

CREATE TABLE participation_and_roles.participations (
    hackathon_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('none', 'individual', 'looking_for_team', 'team_member', 'team_captain')),
    team_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (hackathon_id, user_id)
);

CREATE INDEX idx_participations_user_id ON participation_and_roles.participations(user_id);
CREATE INDEX idx_participations_team_id ON participation_and_roles.participations(team_id) WHERE team_id IS NOT NULL;
CREATE INDEX idx_participations_status ON participation_and_roles.participations(status);

CREATE TABLE participation_and_roles.staff_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hackathon_id UUID NOT NULL,
    target_user_id UUID NOT NULL,
    requested_role VARCHAR(50) NOT NULL CHECK (requested_role IN ('organizer', 'mentor', 'judge')),
    created_by_user_id UUID NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'canceled', 'expired')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_staff_invitations_hackathon_id ON participation_and_roles.staff_invitations(hackathon_id);
CREATE INDEX idx_staff_invitations_target_user_id ON participation_and_roles.staff_invitations(target_user_id);
CREATE INDEX idx_staff_invitations_status ON participation_and_roles.staff_invitations(status);
CREATE INDEX idx_staff_invitations_created_at ON participation_and_roles.staff_invitations(created_at DESC);

CREATE TABLE participation_and_roles.idempotency_keys (
    key TEXT NOT NULL,
    scope TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    response_blob BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (key, scope)
);

CREATE INDEX idx_idempotency_keys_expires_at ON participation_and_roles.idempotency_keys(expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS participation_and_roles.idempotency_keys;
DROP TABLE IF EXISTS participation_and_roles.staff_invitations;
DROP TABLE IF EXISTS participation_and_roles.participations;
DROP TABLE IF EXISTS participation_and_roles.staff_roles;
DROP SCHEMA IF EXISTS participation_and_roles;

-- +goose StatementEnd

