package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type GetMyChatMessagesParams struct {
	HackathonID uuid.UUID
}

type GetMyChatMessagesContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
}

func NewGetMyChatMessagesContext() *GetMyChatMessagesContext {
	return &GetMyChatMessagesContext{}
}

func (c *GetMyChatMessagesContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *GetMyChatMessagesContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *GetMyChatMessagesContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *GetMyChatMessagesContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *GetMyChatMessagesContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *GetMyChatMessagesContext) HackathonStage() string {
	return c.hackathonStage
}

type GetMyChatMessagesPolicy struct{}

func NewGetMyChatMessagesPolicy() *GetMyChatMessagesPolicy {
	return &GetMyChatMessagesPolicy{}
}

func (p *GetMyChatMessagesPolicy) Action() policy.Action {
	return ActionGetMyChatMessages
}

func (p *GetMyChatMessagesPolicy) LoadContext(ctx context.Context, params GetMyChatMessagesParams) (*GetMyChatMessagesContext, error) {
	return NewGetMyChatMessagesContext(), nil
}

func (p *GetMyChatMessagesPolicy) Check(ctx context.Context, pctx *GetMyChatMessagesContext) *policy.Decision {
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
