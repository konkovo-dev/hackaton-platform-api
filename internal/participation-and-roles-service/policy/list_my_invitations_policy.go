package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ListMyInvitationsParams struct{}

type ListMyInvitationsContext struct {
	authenticated bool
	actorUserID   uuid.UUID
}

func NewListMyInvitationsContext() *ListMyInvitationsContext {
	return &ListMyInvitationsContext{}
}

func (c *ListMyInvitationsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *ListMyInvitationsContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *ListMyInvitationsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *ListMyInvitationsContext) SetActorUserID(id uuid.UUID) {
	c.actorUserID = id
}

type ListMyInvitationsPolicy struct {
	policy.BasePolicy
}

func NewListMyInvitationsPolicy() *ListMyInvitationsPolicy {
	return &ListMyInvitationsPolicy{
		BasePolicy: policy.NewBasePolicy(ActionListMyInvitations),
	}
}

func (p *ListMyInvitationsPolicy) LoadContext(ctx context.Context, params ListMyInvitationsParams) (policy.PolicyContext, error) {
	pctx := NewListMyInvitationsContext()

	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return pctx, nil
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return pctx, nil
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)

	return pctx, nil
}

func (p *ListMyInvitationsPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	listCtx := pctx.(*ListMyInvitationsContext)

	if !listCtx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	return decision
}

var _ policy.Policy[ListMyInvitationsParams] = (*ListMyInvitationsPolicy)(nil)
