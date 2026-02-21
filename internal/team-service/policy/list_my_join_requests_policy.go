package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ListMyJoinRequestsParams struct{}

type ListMyJoinRequestsContext struct {
	authenticated bool
	actorUserID   uuid.UUID
}

func NewListMyJoinRequestsContext() *ListMyJoinRequestsContext {
	return &ListMyJoinRequestsContext{}
}

func (c *ListMyJoinRequestsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *ListMyJoinRequestsContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *ListMyJoinRequestsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *ListMyJoinRequestsContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

type ListMyJoinRequestsPolicy struct{}

func NewListMyJoinRequestsPolicy() *ListMyJoinRequestsPolicy {
	return &ListMyJoinRequestsPolicy{}
}

func (p *ListMyJoinRequestsPolicy) Action() policy.Action {
	return ActionListMyJoinRequests
}

func (p *ListMyJoinRequestsPolicy) LoadContext(ctx context.Context, params ListMyJoinRequestsParams) (*ListMyJoinRequestsContext, error) {
	return NewListMyJoinRequestsContext(), nil
}

func (p *ListMyJoinRequestsPolicy) Check(ctx context.Context, pctx *ListMyJoinRequestsContext) *policy.Decision {
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
