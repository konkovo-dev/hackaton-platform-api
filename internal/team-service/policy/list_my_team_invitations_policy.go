package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ListMyTeamInvitationsParams struct{}

type ListMyTeamInvitationsContext struct {
	authenticated bool
	actorUserID   uuid.UUID
}

func NewListMyTeamInvitationsContext() *ListMyTeamInvitationsContext {
	return &ListMyTeamInvitationsContext{}
}

func (c *ListMyTeamInvitationsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *ListMyTeamInvitationsContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *ListMyTeamInvitationsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *ListMyTeamInvitationsContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

type ListMyTeamInvitationsPolicy struct{}

func NewListMyTeamInvitationsPolicy() *ListMyTeamInvitationsPolicy {
	return &ListMyTeamInvitationsPolicy{}
}

func (p *ListMyTeamInvitationsPolicy) Action() policy.Action {
	return ActionListMyTeamInvitations
}

func (p *ListMyTeamInvitationsPolicy) LoadContext(ctx context.Context, params ListMyTeamInvitationsParams) (*ListMyTeamInvitationsContext, error) {
	return NewListMyTeamInvitationsContext(), nil
}

func (p *ListMyTeamInvitationsPolicy) Check(ctx context.Context, pctx *ListMyTeamInvitationsContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	decision.Allow()
	return decision
}
