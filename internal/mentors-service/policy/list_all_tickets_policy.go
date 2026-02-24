package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ListAllTicketsParams struct {
	HackathonID uuid.UUID
}

type ListAllTicketsContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
}

func NewListAllTicketsContext() *ListAllTicketsContext {
	return &ListAllTicketsContext{}
}

func (c *ListAllTicketsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *ListAllTicketsContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *ListAllTicketsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *ListAllTicketsContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *ListAllTicketsContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *ListAllTicketsContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *ListAllTicketsContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *ListAllTicketsContext) ActorRoles() []string {
	return c.actorRoles
}

func (c *ListAllTicketsContext) IsOrganizer() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleOrganizer || role == domain.RoleOwner {
			return true
		}
	}
	return false
}

type ListAllTicketsPolicy struct{}

func NewListAllTicketsPolicy() *ListAllTicketsPolicy {
	return &ListAllTicketsPolicy{}
}

func (p *ListAllTicketsPolicy) Action() policy.Action {
	return ActionListAllTickets
}

func (p *ListAllTicketsPolicy) LoadContext(ctx context.Context, params ListAllTicketsParams) (*ListAllTicketsContext, error) {
	return NewListAllTicketsContext(), nil
}

func (p *ListAllTicketsPolicy) Check(ctx context.Context, pctx *ListAllTicketsContext) *policy.Decision {
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

	if !pctx.IsOrganizer() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only organizers can list all tickets",
		})
		return decision
	}

	decision.Allow()
	return decision
}
