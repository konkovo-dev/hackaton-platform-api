package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type CloseTicketParams struct {
	HackathonID uuid.UUID
	TicketID    uuid.UUID
	Ticket      *entity.Ticket
}

type CloseTicketContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
	ticket         *entity.Ticket
}

func NewCloseTicketContext() *CloseTicketContext {
	return &CloseTicketContext{}
}

func (c *CloseTicketContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *CloseTicketContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *CloseTicketContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *CloseTicketContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *CloseTicketContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *CloseTicketContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *CloseTicketContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *CloseTicketContext) ActorRoles() []string {
	return c.actorRoles
}

func (c *CloseTicketContext) SetTicket(ticket *entity.Ticket) {
	c.ticket = ticket
}

func (c *CloseTicketContext) Ticket() *entity.Ticket {
	return c.ticket
}

func (c *CloseTicketContext) IsOrganizer() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleOrganizer || role == domain.RoleOwner {
			return true
		}
	}
	return false
}

func (c *CloseTicketContext) IsAssignedMentor() bool {
	if c.ticket == nil || c.ticket.AssignedMentorUserID == nil {
		return false
	}
	return *c.ticket.AssignedMentorUserID == c.actorUserID
}

type CloseTicketPolicy struct{}

func NewCloseTicketPolicy() *CloseTicketPolicy {
	return &CloseTicketPolicy{}
}

func (p *CloseTicketPolicy) Action() policy.Action {
	return ActionCloseTicket
}

func (p *CloseTicketPolicy) LoadContext(ctx context.Context, params CloseTicketParams) (*CloseTicketContext, error) {
	pctx := NewCloseTicketContext()
	pctx.SetTicket(params.Ticket)
	return pctx, nil
}

func (p *CloseTicketPolicy) Check(ctx context.Context, pctx *CloseTicketContext) *policy.Decision {
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
			Message: "ticket is already closed",
		})
		return decision
	}

	if !pctx.IsAssignedMentor() && !pctx.IsOrganizer() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only assigned mentor or organizer can close ticket",
		})
		return decision
	}

	decision.Allow()
	return decision
}
