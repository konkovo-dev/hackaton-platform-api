package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type NoParams struct{}

type RequireAuthPolicy struct {
	policy.BasePolicy
}

func NewRequireAuthPolicy(action policy.Action) *RequireAuthPolicy {
	return &RequireAuthPolicy{
		BasePolicy: policy.NewBasePolicy(action),
	}
}

func (p *RequireAuthPolicy) LoadContext(ctx context.Context, params NoParams) (policy.PolicyContext, error) {
	pctx := NewIdentityPolicyContext()

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

func (p *RequireAuthPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	ictx := pctx.(*IdentityPolicyContext)

	if !ictx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
	}

	return decision
}

var _ policy.Policy[NoParams] = (*RequireAuthPolicy)(nil)
