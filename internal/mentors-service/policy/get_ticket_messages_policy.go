package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type GetTicketMessagesParams struct {
	HackathonID uuid.UUID
	TicketID    uuid.UUID
	Ticket      *entity.Ticket
}

type GetTicketMessagesContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
	ticket         *entity.Ticket
	actorTeamID    *uuid.UUID
}

func NewGetTicketMessagesContext() *GetTicketMessagesContext {
	return &GetTicketMessagesContext{}
}

func (c *GetTicketMessagesContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *GetTicketMessagesContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *GetTicketMessagesContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *GetTicketMessagesContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *GetTicketMessagesContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *GetTicketMessagesContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *GetTicketMessagesContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *GetTicketMessagesContext) ActorRoles() []string {
	return c.actorRoles
}

func (c *GetTicketMessagesContext) SetTicket(ticket *entity.Ticket) {
	c.ticket = ticket
}

func (c *GetTicketMessagesContext) Ticket() *entity.Ticket {
	return c.ticket
}

func (c *GetTicketMessagesContext) SetActorTeamID(teamID *uuid.UUID) {
	c.actorTeamID = teamID
}

func (c *GetTicketMessagesContext) ActorTeamID() *uuid.UUID {
	return c.actorTeamID
}

func (c *GetTicketMessagesContext) IsOrganizer() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleOrganizer || role == domain.RoleOwner {
			return true
		}
	}
	return false
}

func (c *GetTicketMessagesContext) IsMentor() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleMentor {
			return true
		}
	}
	return false
}

func (c *GetTicketMessagesContext) IsAssignedMentor() bool {
	if c.ticket == nil || c.ticket.AssignedMentorUserID == nil {
		return false
	}
	return *c.ticket.AssignedMentorUserID == c.actorUserID
}

func (c *GetTicketMessagesContext) IsOwner() bool {
	if c.ticket == nil {
		return false
	}
	
	if c.ticket.OwnerKind == domain.OwnerKindUser {
		return c.ticket.OwnerID == c.actorUserID
	}
	
	if c.ticket.OwnerKind == domain.OwnerKindTeam {
		if c.actorTeamID == nil {
			return false
		}
		return c.ticket.OwnerID == *c.actorTeamID
	}
	
	return false
}

type GetTicketMessagesPolicy struct{}

func NewGetTicketMessagesPolicy() *GetTicketMessagesPolicy {
	return &GetTicketMessagesPolicy{}
}

func (p *GetTicketMessagesPolicy) Action() policy.Action {
	return ActionGetTicketMessages
}

func (p *GetTicketMessagesPolicy) LoadContext(ctx context.Context, params GetTicketMessagesParams) (*GetTicketMessagesContext, error) {
	pctx := NewGetTicketMessagesContext()
	pctx.SetTicket(params.Ticket)
	return pctx, nil
}

func (p *GetTicketMessagesPolicy) Check(ctx context.Context, pctx *GetTicketMessagesContext) *policy.Decision {
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

	if !pctx.IsOwner() && !pctx.IsAssignedMentor() && !pctx.IsOrganizer() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only ticket owner, assigned mentor, or organizers can view messages",
		})
		return decision
	}

	decision.Allow()
	return decision
}
