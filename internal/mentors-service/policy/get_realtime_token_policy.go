package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type GetRealtimeTokenParams struct {
	HackathonID uuid.UUID
}

type GetRealtimeTokenContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
	participates   bool
}

func NewGetRealtimeTokenContext() *GetRealtimeTokenContext {
	return &GetRealtimeTokenContext{}
}

func (c *GetRealtimeTokenContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *GetRealtimeTokenContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *GetRealtimeTokenContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *GetRealtimeTokenContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *GetRealtimeTokenContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *GetRealtimeTokenContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *GetRealtimeTokenContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *GetRealtimeTokenContext) ActorRoles() []string {
	return c.actorRoles
}

func (c *GetRealtimeTokenContext) SetParticipates(participates bool) {
	c.participates = participates
}

func (c *GetRealtimeTokenContext) Participates() bool {
	return c.participates
}

type GetRealtimeTokenPolicy struct{}

func NewGetRealtimeTokenPolicy() *GetRealtimeTokenPolicy {
	return &GetRealtimeTokenPolicy{}
}

func (p *GetRealtimeTokenPolicy) Action() policy.Action {
	return ActionGetRealtimeToken
}

func (p *GetRealtimeTokenPolicy) LoadContext(ctx context.Context, params GetRealtimeTokenParams) (*GetRealtimeTokenContext, error) {
	return NewGetRealtimeTokenContext(), nil
}

func (p *GetRealtimeTokenPolicy) Check(ctx context.Context, pctx *GetRealtimeTokenContext) *policy.Decision {
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
			Message: "realtime token can only be obtained during RUNNING stage",
		})
		return decision
	}

	if !pctx.Participates() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be a participant or have a role in the hackathon",
		})
		return decision
	}

	decision.Allow()
	return decision
}
