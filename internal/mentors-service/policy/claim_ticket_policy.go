package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ClaimTicketParams struct {
	HackathonID uuid.UUID
	TicketID    uuid.UUID
	Ticket      *entity.Ticket
}

type ClaimTicketContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
	ticket         *entity.Ticket
}

func NewClaimTicketContext() *ClaimTicketContext {
	return &ClaimTicketContext{}
}

func (c *ClaimTicketContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *ClaimTicketContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *ClaimTicketContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *ClaimTicketContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *ClaimTicketContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *ClaimTicketContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *ClaimTicketContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *ClaimTicketContext) ActorRoles() []string {
	return c.actorRoles
}

func (c *ClaimTicketContext) SetTicket(ticket *entity.Ticket) {
	c.ticket = ticket
}

func (c *ClaimTicketContext) Ticket() *entity.Ticket {
	return c.ticket
}

func (c *ClaimTicketContext) IsMentor() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleMentor {
			return true
		}
	}
	return false
}

type ClaimTicketPolicy struct{}

func NewClaimTicketPolicy() *ClaimTicketPolicy {
	return &ClaimTicketPolicy{}
}

func (p *ClaimTicketPolicy) Action() policy.Action {
	return ActionClaimTicket
}

func (p *ClaimTicketPolicy) LoadContext(ctx context.Context, params ClaimTicketParams) (*ClaimTicketContext, error) {
	pctx := NewClaimTicketContext()
	pctx.SetTicket(params.Ticket)
	return pctx, nil
}

func (p *ClaimTicketPolicy) Check(ctx context.Context, pctx *ClaimTicketContext) *policy.Decision {
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
			Message: "tickets can only be claimed during RUNNING stage",
		})
		return decision
	}

	if !pctx.IsMentor() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only mentors can claim tickets",
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

	if pctx.Ticket().AssignedMentorUserID != nil {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "ticket is already assigned",
		})
		return decision
	}

	decision.Allow()
	return decision
}
