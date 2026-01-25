package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type UpdateMeParams struct {
	UserID    uuid.UUID
	FirstName string
	LastName  string
	Timezone  string
}

type UpdateMePolicy struct {
	policy.BasePolicy
}

func NewUpdateMePolicy() *UpdateMePolicy {
	return &UpdateMePolicy{
		BasePolicy: policy.NewBasePolicy(ActionUpdateMe),
	}
}

func (p *UpdateMePolicy) LoadContext(ctx context.Context, params UpdateMeParams) (policy.PolicyContext, error) {
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

func (p *UpdateMePolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
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

func (p *UpdateMePolicy) ValidateInput(params UpdateMeParams) *policy.Decision {
	decision := policy.NewDecision()

	if params.FirstName == "" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeRequired,
			Field:   "first_name",
			Message: "first_name is required",
		})
	}

	if params.LastName == "" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeRequired,
			Field:   "last_name",
			Message: "last_name is required",
		})
	}

	if params.Timezone == "" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeRequired,
			Field:   "timezone",
			Message: "timezone is required",
		})
	}

	return decision
}

var _ policy.Policy[UpdateMeParams] = (*UpdateMePolicy)(nil)
