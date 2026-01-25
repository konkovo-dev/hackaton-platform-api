package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
)

type LogoutParams struct {
	RefreshToken string
}

type LogoutPolicy struct {
	policy.BasePolicy
}

func NewLogoutPolicy() *LogoutPolicy {
	return &LogoutPolicy{
		BasePolicy: policy.NewBasePolicy(ActionLogout),
	}
}

func (p *LogoutPolicy) LoadContext(ctx context.Context, params LogoutParams) (policy.PolicyContext, error) {
	return NewAuthPolicyContext(), nil
}

func (p *LogoutPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	return decision
}

func (p *LogoutPolicy) ValidateInput(params LogoutParams) *policy.Decision {
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

var _ policy.Policy[LogoutParams] = (*LogoutPolicy)(nil)
