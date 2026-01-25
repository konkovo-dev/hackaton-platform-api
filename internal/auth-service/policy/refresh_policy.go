package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
)

type RefreshParams struct {
	RefreshToken string
}

type RefreshPolicy struct {
	policy.BasePolicy
}

func NewRefreshPolicy() *RefreshPolicy {
	return &RefreshPolicy{
		BasePolicy: policy.NewBasePolicy(ActionRefresh),
	}
}

func (p *RefreshPolicy) LoadContext(ctx context.Context, params RefreshParams) (policy.PolicyContext, error) {
	return NewAuthPolicyContext(), nil
}

func (p *RefreshPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	return decision
}

func (p *RefreshPolicy) ValidateInput(params RefreshParams) *policy.Decision {
	decision := policy.NewDecision()

	if params.RefreshToken == "" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeRequired,
			Field:   "refresh_token",
			Message: "refresh_token is required",
		})
	}

	return decision
}

var _ policy.Policy[RefreshParams] = (*RefreshPolicy)(nil)
