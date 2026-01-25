package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type PublishHackathonParams struct {
	HackathonID uuid.UUID
}

type PublishHackathonPolicy struct {
	policy.BasePolicy
	hackathonRepo HackathonRepository
	parClient     ParticipationAndRolesClient
}

func NewPublishHackathonPolicy(hackathonRepo HackathonRepository, parClient ParticipationAndRolesClient) *PublishHackathonPolicy {
	return &PublishHackathonPolicy{
		BasePolicy:    policy.NewBasePolicy(ActionPublishHackathon),
		hackathonRepo: hackathonRepo,
		parClient:     parClient,
	}
}

func (p *PublishHackathonPolicy) LoadContext(ctx context.Context, params PublishHackathonParams) (policy.PolicyContext, error) {
	return loadHackathonPolicyContext(ctx, params.HackathonID, p.hackathonRepo, p.parClient)
}

func (p *PublishHackathonPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	hctx := pctx.(*HackathonPolicyContext)

	if !hctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	if hctx.Stage() != string(domain.StageDraft) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only draft hackathons can be published",
		})
		return decision
	}

	if hctx.PublishedAt() != nil {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "hackathon already published",
		})
		return decision
	}

	hasOwnerRole := hctx.HasRole(string(domain.RoleOwner))
	if !hasOwnerRole {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only owner can publish hackathon",
		})
	}

	return decision
}

var _ policy.Policy[PublishHackathonParams] = (*PublishHackathonPolicy)(nil)
