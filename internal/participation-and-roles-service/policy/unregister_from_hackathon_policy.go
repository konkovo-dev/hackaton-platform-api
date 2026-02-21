package policy

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type UnregisterFromHackathonParams struct {
	HackathonID uuid.UUID
}

type UnregisterFromHackathonContext struct {
	authenticated bool
	actorUserID   uuid.UUID

	participationStatus string
}

func NewUnregisterFromHackathonContext() *UnregisterFromHackathonContext {
	return &UnregisterFromHackathonContext{}
}

func (c *UnregisterFromHackathonContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *UnregisterFromHackathonContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *UnregisterFromHackathonContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *UnregisterFromHackathonContext) SetActorUserID(id uuid.UUID) {
	c.actorUserID = id
}

func (c *UnregisterFromHackathonContext) SetParticipationStatus(status string) {
	c.participationStatus = status
}

func (c *UnregisterFromHackathonContext) CanUnregister() bool {
	return c.participationStatus != string(domain.ParticipationTeamMember) &&
		c.participationStatus != string(domain.ParticipationTeamCaptain)
}

func (c *UnregisterFromHackathonContext) ParticipationStatus() string {
	return c.participationStatus
}

type UnregisterFromHackathonRepositories interface {
	GetParticipationStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error)
}

type UnregisterFromHackathonPolicy struct {
	repos UnregisterFromHackathonRepositories
}

func NewUnregisterFromHackathonPolicy(repos UnregisterFromHackathonRepositories) *UnregisterFromHackathonPolicy {
	return &UnregisterFromHackathonPolicy{
		repos: repos,
	}
}

func (p *UnregisterFromHackathonPolicy) LoadContext(ctx context.Context, params UnregisterFromHackathonParams) (*UnregisterFromHackathonContext, error) {
	pctx := NewUnregisterFromHackathonContext()

	userID, ok := auth.GetUserID(ctx)
	if !ok {
		pctx.SetAuthenticated(false)
		return pctx, nil
	}

	pctx.SetAuthenticated(true)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user id: %w", err)
	}

	pctx.SetActorUserID(userUUID)

	participationStatus, err := p.repos.GetParticipationStatus(ctx, params.HackathonID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participation status: %w", err)
	}

	pctx.SetParticipationStatus(participationStatus)

	return pctx, nil
}

func (p *UnregisterFromHackathonPolicy) Check(ctx context.Context, pctx *UnregisterFromHackathonContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if pctx.ParticipationStatus() == "" || pctx.ParticipationStatus() == string(domain.ParticipationNone) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeNotFound,
			Message: "participation not found",
		})
		return decision
	}

	if !pctx.CanUnregister() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "cannot unregister while being in a team",
		})
		return decision
	}

	decision.Allow()
	return decision
}
