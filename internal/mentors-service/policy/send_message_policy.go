package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type SendMessageParams struct {
	HackathonID uuid.UUID
}

type SendMessageContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
}

func NewSendMessageContext() *SendMessageContext {
	return &SendMessageContext{}
}

func (c *SendMessageContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *SendMessageContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *SendMessageContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *SendMessageContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *SendMessageContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *SendMessageContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *SendMessageContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *SendMessageContext) ActorRoles() []string {
	return c.actorRoles
}

func (c *SendMessageContext) IsStaff() bool {
	for _, role := range c.actorRoles {
		if role == "owner" || role == "organizer" || role == "mentor" {
			return true
		}
	}
	return false
}

type SendMessagePolicy struct{}

func NewSendMessagePolicy() *SendMessagePolicy {
	return &SendMessagePolicy{}
}

func (p *SendMessagePolicy) Action() policy.Action {
	return ActionSendMessage
}

func (p *SendMessagePolicy) LoadContext(ctx context.Context, params SendMessageParams) (*SendMessageContext, error) {
	return NewSendMessageContext(), nil
}

func (p *SendMessagePolicy) Check(ctx context.Context, pctx *SendMessageContext) *policy.Decision {
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

	if pctx.IsStaff() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "staff cannot send messages as participants",
		})
		return decision
	}

	decision.Allow()
	return decision
}
