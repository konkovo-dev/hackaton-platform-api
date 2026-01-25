package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ReadTaskParams struct {
	HackathonID uuid.UUID
}

type ReadTaskPolicy struct {
	policy.BasePolicy
	hackathonRepo HackathonRepository
	parClient     ParticipationAndRolesClient
}

func NewReadTaskPolicy(hackathonRepo HackathonRepository, parClient ParticipationAndRolesClient) *ReadTaskPolicy {
	return &ReadTaskPolicy{
		BasePolicy:    policy.NewBasePolicy(ActionReadTask),
		hackathonRepo: hackathonRepo,
		parClient:     parClient,
	}
}

func (p *ReadTaskPolicy) LoadContext(ctx context.Context, params ReadTaskParams) (policy.PolicyContext, error) {
	return loadHackathonPolicyContext(ctx, params.HackathonID, p.hackathonRepo, p.parClient)
}

func (p *ReadTaskPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
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
	hasMentorRole := hctx.HasRole(string(domain.RoleMentor))
	hasJuryRole := hctx.HasRole(string(domain.RoleJury))

	stage := domain.HackathonStage(hctx.Stage())
	participationKind := domain.ParticipationStatus(hctx.ParticipationKind())

	if hasOwnerRole || hasOrganizerRole {
		return decision
	}

	if stage != domain.StageDraft && (hasMentorRole || hasJuryRole) {
		return decision
	}

	if stage == domain.StageRunning && (participationKind == domain.ParticipationIndividual || participationKind == domain.ParticipationTeamMember || participationKind == domain.ParticipationTeamCaptain) {
		return decision
	}

	decision.Deny(policy.Violation{
		Code:    policy.ViolationCodeForbidden,
		Message: "insufficient permissions to read task",
	})
	return decision
}

var _ policy.Policy[ReadTaskParams] = (*ReadTaskPolicy)(nil)

type UpdateTaskParams struct {
	HackathonID uuid.UUID
	NewTask     string
}

type UpdateTaskPolicy struct {
	policy.BasePolicy
	hackathonRepo HackathonRepository
	parClient     ParticipationAndRolesClient
}

func NewUpdateTaskPolicy(hackathonRepo HackathonRepository, parClient ParticipationAndRolesClient) *UpdateTaskPolicy {
	return &UpdateTaskPolicy{
		BasePolicy:    policy.NewBasePolicy(ActionUpdateTask),
		hackathonRepo: hackathonRepo,
		parClient:     parClient,
	}
}

func (p *UpdateTaskPolicy) LoadContext(ctx context.Context, params UpdateTaskParams) (policy.PolicyContext, error) {
	return loadHackathonPolicyContext(ctx, params.HackathonID, p.hackathonRepo, p.parClient)
}

func (p *UpdateTaskPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
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
			Message: "only owners and organizers can update task",
		})
		return decision
	}

	stage := domain.HackathonStage(hctx.Stage())
	if stage == domain.StageJudging || stage == domain.StageFinished {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "task cannot be updated during JUDGING or FINISHED stages",
		})
		return decision
	}

	return decision
}

var _ policy.Policy[UpdateTaskParams] = (*UpdateTaskPolicy)(nil)
