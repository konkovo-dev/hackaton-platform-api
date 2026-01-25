package policy

import (
	"context"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type UpdateHackathonParams struct {
	HackathonID             uuid.UUID
	NewName                 string
	NewLocationOnline       bool
	NewLocationCity         string
	NewLocationCountry      string
	NewRegistrationOpensAt  *time.Time
	NewRegistrationClosesAt *time.Time
	NewStartsAt             *time.Time
	NewEndsAt               *time.Time
	NewJudgingEndsAt        *time.Time
	NewTeamSizeMax          int32
	NewAllowIndividual      bool
	NewAllowTeam            bool
}

type UpdateHackathonPolicy struct {
	policy.BasePolicy
	hackathonRepo HackathonRepository
	parClient     ParticipationAndRolesClient
}

func NewUpdateHackathonPolicy(hackathonRepo HackathonRepository, parClient ParticipationAndRolesClient) *UpdateHackathonPolicy {
	return &UpdateHackathonPolicy{
		BasePolicy:    policy.NewBasePolicy(ActionUpdateHackathon),
		hackathonRepo: hackathonRepo,
		parClient:     parClient,
	}
}

func (p *UpdateHackathonPolicy) LoadContext(ctx context.Context, params UpdateHackathonParams) (policy.PolicyContext, error) {
	return loadHackathonPolicyContext(ctx, params.HackathonID, p.hackathonRepo, p.parClient)
}

func (p *UpdateHackathonPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
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
			Message: "only owners and organizers can update hackathon",
		})
	}

	return decision
}

var _ policy.Policy[UpdateHackathonParams] = (*UpdateHackathonPolicy)(nil)
