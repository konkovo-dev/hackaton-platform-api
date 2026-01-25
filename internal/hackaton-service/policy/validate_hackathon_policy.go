package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ValidateHackathonParams struct {
	HackathonID uuid.UUID
}

type ValidateHackathonPolicy struct {
	policy.BasePolicy
	hackathonRepo HackathonRepository
	parClient     ParticipationAndRolesClient
}

func NewValidateHackathonPolicy(hackathonRepo HackathonRepository, parClient ParticipationAndRolesClient) *ValidateHackathonPolicy {
	return &ValidateHackathonPolicy{
		BasePolicy:    policy.NewBasePolicy(ActionValidateHackathon),
		hackathonRepo: hackathonRepo,
		parClient:     parClient,
	}
}

func (p *ValidateHackathonPolicy) LoadContext(ctx context.Context, params ValidateHackathonParams) (policy.PolicyContext, error) {
	return loadHackathonPolicyContext(ctx, params.HackathonID, p.hackathonRepo, p.parClient)
}

func (p *ValidateHackathonPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	hctx := pctx.(*HackathonPolicyContext)

	if !hctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	hasOwnerRole := hctx.HasRole(string(domain.RoleOwner))
	hasOrganizerRole := hctx.HasRole(string(domain.RoleOrganizer))

	if !hasOwnerRole && !hasOrganizerRole {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only owners and organizers can validate hackathon",
		})
	}

	return decision
}

var _ policy.Policy[ValidateHackathonParams] = (*ValidateHackathonPolicy)(nil)
