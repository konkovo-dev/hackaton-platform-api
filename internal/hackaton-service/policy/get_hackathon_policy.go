package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type HackathonRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Hackathon, error)
}

type GetHackathonParams struct {
	HackathonID uuid.UUID
}

type ParticipationAndRolesClient interface {
	GetHackathonContext(ctx context.Context, hackathonID string) (userID, participationStatus, teamID string, roles []string, err error)
}

type GetHackathonPolicy struct {
	policy.BasePolicy
	hackathonRepo HackathonRepository
	parClient     ParticipationAndRolesClient
}

func NewGetHackathonPolicy(hackathonRepo HackathonRepository, parClient ParticipationAndRolesClient) *GetHackathonPolicy {
	return &GetHackathonPolicy{
		BasePolicy:    policy.NewBasePolicy(ActionReadHackathon),
		hackathonRepo: hackathonRepo,
		parClient:     parClient,
	}
}

func (p *GetHackathonPolicy) LoadContext(ctx context.Context, params GetHackathonParams) (policy.PolicyContext, error) {
	return loadHackathonPolicyContext(ctx, params.HackathonID, p.hackathonRepo, p.parClient)
}

func (p *GetHackathonPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	hctx := pctx.(*HackathonPolicyContext)

	if hctx.State() == string(domain.StateDraft) {
		hasOwnerRole := hctx.HasRole(string(domain.RoleOwner))
		hasOrganizerRole := hctx.HasRole(string(domain.RoleOrganizer))

		if !hasOwnerRole && !hasOrganizerRole {
			decision.Deny(policy.Violation{
				Code:    policy.ViolationCodeForbidden,
				Message: "draft hackathons are only visible to owners and organizers",
			})
		}
	}

	return decision
}

var _ policy.Policy[GetHackathonParams] = (*GetHackathonPolicy)(nil)
