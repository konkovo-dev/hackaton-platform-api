package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type UpdateMyContactsParams struct {
	UserID uuid.UUID
}

type UpdateMyContactsPolicy struct {
	policy.BasePolicy
}

func NewUpdateMyContactsPolicy() *UpdateMyContactsPolicy {
	return &UpdateMyContactsPolicy{
		BasePolicy: policy.NewBasePolicy(ActionUpdateMyContacts),
	}
}

func (p *UpdateMyContactsPolicy) LoadContext(ctx context.Context, params UpdateMyContactsParams) (policy.PolicyContext, error) {
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

func (p *UpdateMyContactsPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
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

var _ policy.Policy[UpdateMyContactsParams] = (*UpdateMyContactsPolicy)(nil)
