package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ReadMeParams struct {
	UserID uuid.UUID
}

type ReadMePolicy struct {
	policy.BasePolicy
}

func NewReadMePolicy() *ReadMePolicy {
	return &ReadMePolicy{
		BasePolicy: policy.NewBasePolicy(ActionReadMe),
	}
}

func (p *ReadMePolicy) LoadContext(ctx context.Context, params ReadMeParams) (policy.PolicyContext, error) {
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
	pctx.SetTargetUserID(userUUID)
	pctx.SetIsMe(true)

	return pctx, nil
}

func (p *ReadMePolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	ictx := pctx.(*IdentityPolicyContext)

	if !ictx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	if !ictx.IsMe() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "access denied: not your profile",
		})
		return decision
	}

	return decision
}

var _ policy.Policy[ReadMeParams] = (*ReadMePolicy)(nil)
