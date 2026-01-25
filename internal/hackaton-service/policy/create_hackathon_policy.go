package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type CreateHackathonParams struct{}

type CreateHackathonPolicy struct {
	policy.BasePolicy
}

func NewCreateHackathonPolicy() *CreateHackathonPolicy {
	return &CreateHackathonPolicy{
		BasePolicy: policy.NewBasePolicy(ActionCreateHackathon),
	}
}

func (p *CreateHackathonPolicy) LoadContext(ctx context.Context, params CreateHackathonParams) (policy.PolicyContext, error) {
	pctx := NewHackathonPolicyContext()

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

func (p *CreateHackathonPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	hctx := pctx.(*HackathonPolicyContext)

	if !hctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	return decision
}

var _ policy.Policy[CreateHackathonParams] = (*CreateHackathonPolicy)(nil)
