package policy

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type UpdateMyParticipationParams struct {
	HackathonID uuid.UUID
}

type UpdateMyParticipationContext struct {
	authenticated bool
	actorUserID   uuid.UUID

	participationStatus string
}

func NewUpdateMyParticipationContext() *UpdateMyParticipationContext {
	return &UpdateMyParticipationContext{}
}

func (c *UpdateMyParticipationContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *UpdateMyParticipationContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *UpdateMyParticipationContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *UpdateMyParticipationContext) SetActorUserID(id uuid.UUID) {
	c.actorUserID = id
}

func (c *UpdateMyParticipationContext) SetParticipationStatus(status string) {
	c.participationStatus = status
}

func (c *UpdateMyParticipationContext) CanUpdate() bool {
	return c.participationStatus == string(domain.ParticipationIndividual) ||
		c.participationStatus == string(domain.ParticipationLookingForTeam)
}

func (c *UpdateMyParticipationContext) ParticipationStatus() string {
	return c.participationStatus
}

type UpdateMyParticipationRepositories interface {
	GetParticipationStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error)
}

type UpdateMyParticipationPolicy struct {
	repos UpdateMyParticipationRepositories
}

func NewUpdateMyParticipationPolicy(repos UpdateMyParticipationRepositories) *UpdateMyParticipationPolicy {
	return &UpdateMyParticipationPolicy{
		repos: repos,
	}
}

func (p *UpdateMyParticipationPolicy) LoadContext(ctx context.Context, params UpdateMyParticipationParams) (*UpdateMyParticipationContext, error) {
	pctx := NewUpdateMyParticipationContext()

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

func (p *UpdateMyParticipationPolicy) Check(ctx context.Context, pctx *UpdateMyParticipationContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if !pctx.CanUpdate() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "can only update participation in INDIVIDUAL_ACTIVE or LOOKING_FOR_TEAM status",
		})
		return decision
	}

	decision.Allow()
	return decision
}
