package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type GetMyTicketsParams struct {
	HackathonID uuid.UUID
}

type GetMyTicketsContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
}

func NewGetMyTicketsContext() *GetMyTicketsContext {
	return &GetMyTicketsContext{}
}

func (c *GetMyTicketsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *GetMyTicketsContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *GetMyTicketsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *GetMyTicketsContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *GetMyTicketsContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *GetMyTicketsContext) HackathonStage() string {
	return c.hackathonStage
}

type GetMyTicketsPolicy struct{}

func NewGetMyTicketsPolicy() *GetMyTicketsPolicy {
	return &GetMyTicketsPolicy{}
}

func (p *GetMyTicketsPolicy) Action() policy.Action {
	return ActionGetMyTickets
}

func (p *GetMyTicketsPolicy) LoadContext(ctx context.Context, params GetMyTicketsParams) (*GetMyTicketsContext, error) {
	return NewGetMyTicketsContext(), nil
}

func (p *GetMyTicketsPolicy) Check(ctx context.Context, pctx *GetMyTicketsContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if pctx.HackathonStage() != domain.HackathonStageRunning {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeStageRule,
			Message: "support chat is only available during RUNNING stage",
		})
		return decision
	}

	decision.Allow()
	return decision
}
