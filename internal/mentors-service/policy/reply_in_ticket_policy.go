package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ReplyInTicketParams struct {
	HackathonID uuid.UUID
	TicketID    uuid.UUID
	Ticket      *entity.Ticket
}

type ReplyInTicketContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
	ticket         *entity.Ticket
}

func NewReplyInTicketContext() *ReplyInTicketContext {
	return &ReplyInTicketContext{}
}

func (c *ReplyInTicketContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *ReplyInTicketContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *ReplyInTicketContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *ReplyInTicketContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *ReplyInTicketContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *ReplyInTicketContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *ReplyInTicketContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *ReplyInTicketContext) ActorRoles() []string {
	return c.actorRoles
}

func (c *ReplyInTicketContext) SetTicket(ticket *entity.Ticket) {
	c.ticket = ticket
}

func (c *ReplyInTicketContext) Ticket() *entity.Ticket {
	return c.ticket
}

func (c *ReplyInTicketContext) IsOrganizer() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleOrganizer || role == domain.RoleOwner {
			return true
		}
	}
	return false
}

func (c *ReplyInTicketContext) IsAssignedMentor() bool {
	if c.ticket == nil || c.ticket.AssignedMentorUserID == nil {
		return false
	}
	return *c.ticket.AssignedMentorUserID == c.actorUserID
}

type ReplyInTicketPolicy struct{}

func NewReplyInTicketPolicy() *ReplyInTicketPolicy {
	return &ReplyInTicketPolicy{}
}

func (p *ReplyInTicketPolicy) Action() policy.Action {
	return ActionReplyInTicket
}

func (p *ReplyInTicketPolicy) LoadContext(ctx context.Context, params ReplyInTicketParams) (*ReplyInTicketContext, error) {
	pctx := NewReplyInTicketContext()
	pctx.SetTicket(params.Ticket)
	return pctx, nil
}

func (p *ReplyInTicketPolicy) Check(ctx context.Context, pctx *ReplyInTicketContext) *policy.Decision {
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
			Message: "ticket is not open",
		})
		return decision
	}

	// Only assigned mentor can reply in ticket
	// Participants use SendMessage, organizers have read-only access
	if !pctx.IsAssignedMentor() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only assigned mentor can reply in ticket",
		})
		return decision
	}

	decision.Allow()
	return decision
}
