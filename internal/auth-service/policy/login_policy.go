package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
)

type LoginParams struct {
	Email    string
	Username string
	Password string
}

type LoginPolicy struct {
	policy.BasePolicy
}

func NewLoginPolicy() *LoginPolicy {
	return &LoginPolicy{
		BasePolicy: policy.NewBasePolicy(ActionLogin),
	}
}

func (p *LoginPolicy) LoadContext(ctx context.Context, params LoginParams) (policy.PolicyContext, error) {
	return NewAuthPolicyContext(), nil
}

func (p *LoginPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	return decision
}

func (p *LoginPolicy) ValidateInput(params LoginParams) *policy.Decision {
	decision := policy.NewDecision()

	if params.Email == "" && params.Username == "" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeRequired,
			Field:   "login",
			Message: "email or username is required",
		})
	}

	if params.Password == "" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeRequired,
			Field:   "password",
			Message: "password is required",
		})
	}

	return decision
}

var _ policy.Policy[LoginParams] = (*LoginPolicy)(nil)
