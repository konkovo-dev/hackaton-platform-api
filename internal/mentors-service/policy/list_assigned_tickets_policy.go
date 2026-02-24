package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ListAssignedTicketsParams struct {
	HackathonID uuid.UUID
}

type ListAssignedTicketsContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
}

func NewListAssignedTicketsContext() *ListAssignedTicketsContext {
	return &ListAssignedTicketsContext{}
}

func (c *ListAssignedTicketsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *ListAssignedTicketsContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *ListAssignedTicketsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *ListAssignedTicketsContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *ListAssignedTicketsContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *ListAssignedTicketsContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *ListAssignedTicketsContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *ListAssignedTicketsContext) ActorRoles() []string {
	return c.actorRoles
}

func (c *ListAssignedTicketsContext) IsMentor() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleMentor {
			return true
		}
	}
	return false
}

type ListAssignedTicketsPolicy struct{}

func NewListAssignedTicketsPolicy() *ListAssignedTicketsPolicy {
	return &ListAssignedTicketsPolicy{}
}

func (p *ListAssignedTicketsPolicy) Action() policy.Action {
	return ActionListAssignedTickets
}

func (p *ListAssignedTicketsPolicy) LoadContext(ctx context.Context, params ListAssignedTicketsParams) (*ListAssignedTicketsContext, error) {
	return NewListAssignedTicketsContext(), nil
}

func (p *ListAssignedTicketsPolicy) Check(ctx context.Context, pctx *ListAssignedTicketsContext) *policy.Decision {
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

	if !pctx.IsMentor() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only mentors can list assigned tickets",
		})
		return decision
	}

	decision.Allow()
	return decision
}
