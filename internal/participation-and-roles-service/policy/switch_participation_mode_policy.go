package policy

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type SwitchParticipationModeParams struct {
	HackathonID uuid.UUID
	NewStatus   string
}

type SwitchParticipationModeContext struct {
	authenticated bool
	actorUserID   uuid.UUID

	participationStatus string
	newStatus           string
}

func NewSwitchParticipationModeContext() *SwitchParticipationModeContext {
	return &SwitchParticipationModeContext{}
}

func (c *SwitchParticipationModeContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *SwitchParticipationModeContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *SwitchParticipationModeContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *SwitchParticipationModeContext) SetActorUserID(id uuid.UUID) {
	c.actorUserID = id
}

func (c *SwitchParticipationModeContext) SetParticipationStatus(status string) {
	c.participationStatus = status
}

func (c *SwitchParticipationModeContext) SetNewStatus(status string) {
	c.newStatus = status
}

func (c *SwitchParticipationModeContext) CanSwitch() bool {
	return c.participationStatus == string(domain.ParticipationIndividual) ||
		c.participationStatus == string(domain.ParticipationLookingForTeam)
}

func (c *SwitchParticipationModeContext) IsStatusDifferent() bool {
	return c.participationStatus != c.newStatus
}

func (c *SwitchParticipationModeContext) ParticipationStatus() string {
	return c.participationStatus
}

func (c *SwitchParticipationModeContext) NewStatus() string {
	return c.newStatus
}

type SwitchParticipationModeRepositories interface {
	GetParticipationStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error)
}

type SwitchParticipationModePolicy struct {
	repos SwitchParticipationModeRepositories
}

func NewSwitchParticipationModePolicy(repos SwitchParticipationModeRepositories) *SwitchParticipationModePolicy {
	return &SwitchParticipationModePolicy{
		repos: repos,
	}
}

func (p *SwitchParticipationModePolicy) LoadContext(ctx context.Context, params SwitchParticipationModeParams) (*SwitchParticipationModeContext, error) {
	pctx := NewSwitchParticipationModeContext()

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
	pctx.SetNewStatus(params.NewStatus)

	return pctx, nil
}

func (p *SwitchParticipationModePolicy) Check(ctx context.Context, pctx *SwitchParticipationModeContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if !pctx.CanSwitch() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "can only switch between INDIVIDUAL_ACTIVE and LOOKING_FOR_TEAM",
		})
		return decision
	}

	if !pctx.IsStatusDifferent() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "new status must be different from current status",
		})
		return decision
	}

	decision.Allow()
	return decision
}
