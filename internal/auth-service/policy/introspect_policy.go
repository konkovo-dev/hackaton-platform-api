package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
)

type IntrospectParams struct {
	AccessToken string
}

type IntrospectPolicy struct {
	policy.BasePolicy
}

func NewIntrospectPolicy() *IntrospectPolicy {
	return &IntrospectPolicy{
		BasePolicy: policy.NewBasePolicy(ActionIntrospectToken),
	}
}

func (p *IntrospectPolicy) LoadContext(ctx context.Context, params IntrospectParams) (policy.PolicyContext, error) {
	return NewAuthPolicyContext(), nil
}

func (p *IntrospectPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	return decision
}

func (p *IntrospectPolicy) ValidateInput(params IntrospectParams) *policy.Decision {
	decision := policy.NewDecision()

	if params.AccessToken == "" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeRequired,
			Field:   "access_token",
			Message: "access_token is required",
		})
	}

	return decision
}

var _ policy.Policy[IntrospectParams] = (*IntrospectPolicy)(nil)
