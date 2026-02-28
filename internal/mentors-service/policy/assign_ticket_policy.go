package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type AssignTicketParams struct {
	HackathonID uuid.UUID
	TicketID    uuid.UUID
	Ticket      *entity.Ticket
}

type AssignTicketContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
	ticket         *entity.Ticket
}

func NewAssignTicketContext() *AssignTicketContext {
	return &AssignTicketContext{}
}

func (c *AssignTicketContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *AssignTicketContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *AssignTicketContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *AssignTicketContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *AssignTicketContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *AssignTicketContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *AssignTicketContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *AssignTicketContext) ActorRoles() []string {
	return c.actorRoles
}

func (c *AssignTicketContext) SetTicket(ticket *entity.Ticket) {
	c.ticket = ticket
}

func (c *AssignTicketContext) Ticket() *entity.Ticket {
	return c.ticket
}

func (c *AssignTicketContext) IsOrganizerOrOwner() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleOrganizer || role == domain.RoleOwner {
			return true
		}
	}
	return false
}

type AssignTicketPolicy struct{}

func NewAssignTicketPolicy() *AssignTicketPolicy {
	return &AssignTicketPolicy{}
}

func (p *AssignTicketPolicy) Action() policy.Action {
	return ActionAssignTicket
}

func (p *AssignTicketPolicy) LoadContext(ctx context.Context, params AssignTicketParams) (*AssignTicketContext, error) {
	pctx := NewAssignTicketContext()
	pctx.SetTicket(params.Ticket)
	return pctx, nil
}

func (p *AssignTicketPolicy) Check(ctx context.Context, pctx *AssignTicketContext) *policy.Decision {
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
			Message: "tickets can only be assigned during RUNNING stage",
		})
		return decision
	}

	if !pctx.IsOrganizerOrOwner() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only organizers can assign tickets",
		})
		return decision
	}

	if pctx.Ticket() == nil {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeNotFound,
			Message: "ticket not found",
		})
		return decision
	}

	if pctx.Ticket().Status != domain.TicketStatusOpen {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "ticket is not open",
		})
		return decision
	}

	decision.Allow()
	return decision
}
