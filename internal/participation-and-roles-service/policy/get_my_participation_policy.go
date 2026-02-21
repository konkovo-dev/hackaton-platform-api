package policy

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type GetMyParticipationParams struct {
	HackathonID uuid.UUID
}

type GetMyParticipationContext struct {
	authenticated bool
	actorUserID   uuid.UUID

	participationStatus string
}

func NewGetMyParticipationContext() *GetMyParticipationContext {
	return &GetMyParticipationContext{}
}

func (c *GetMyParticipationContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *GetMyParticipationContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *GetMyParticipationContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *GetMyParticipationContext) SetActorUserID(id uuid.UUID) {
	c.actorUserID = id
}

func (c *GetMyParticipationContext) SetParticipationStatus(status string) {
	c.participationStatus = status
}

func (c *GetMyParticipationContext) IsParticipant() bool {
	return c.participationStatus != "" && c.participationStatus != "NONE"
}

func (c *GetMyParticipationContext) ParticipationStatus() string {
	return c.participationStatus
}

type GetMyParticipationRepositories interface {
	GetParticipationStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error)
}

type GetMyParticipationPolicy struct {
	repos GetMyParticipationRepositories
}

func NewGetMyParticipationPolicy(repos GetMyParticipationRepositories) *GetMyParticipationPolicy {
	return &GetMyParticipationPolicy{
		repos: repos,
	}
}

func (p *GetMyParticipationPolicy) LoadContext(ctx context.Context, params GetMyParticipationParams) (*GetMyParticipationContext, error) {
	pctx := NewGetMyParticipationContext()

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

func (p *GetMyParticipationPolicy) Check(ctx context.Context, pctx *GetMyParticipationContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if !pctx.IsParticipant() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user is not a participant in this hackathon",
		})
		return decision
	}

	decision.Allow()
	return decision
}
