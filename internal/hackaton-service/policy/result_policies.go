package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ReadResultParams struct {
	HackathonID uuid.UUID
}

type ReadResultPolicy struct {
	policy.BasePolicy
	hackathonRepo HackathonRepository
	parClient     ParticipationAndRolesClient
}

func NewReadResultPolicy(hackathonRepo HackathonRepository, parClient ParticipationAndRolesClient) *ReadResultPolicy {
	return &ReadResultPolicy{
		BasePolicy:    policy.NewBasePolicy(ActionReadResult),
		hackathonRepo: hackathonRepo,
		parClient:     parClient,
	}
}

func (p *ReadResultPolicy) LoadContext(ctx context.Context, params ReadResultParams) (policy.PolicyContext, error) {
	return loadHackathonPolicyContext(ctx, params.HackathonID, p.hackathonRepo, p.parClient)
}

func (p *ReadResultPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	hctx := pctx.(*HackathonPolicyContext)

	stage := domain.HackathonStage(hctx.Stage())

	if stage == domain.StageFinished {
		return decision
	}

	if !hctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	hasOwnerRole := hctx.HasRole(string(domain.RoleOwner))
	hasOrganizerRole := hctx.HasRole(string(domain.RoleOrganizer))

	if stage == domain.StageJudging && hctx.ResultPublishedAt() == nil && (hasOwnerRole || hasOrganizerRole) {
		return decision
	}

	decision.Deny(policy.Violation{
		Code:    policy.ViolationCodeForbidden,
		Message: "insufficient permissions to read result",
	})
	return decision
}

var _ policy.Policy[ReadResultParams] = (*ReadResultPolicy)(nil)

type UpdateResultDraftParams struct {
	HackathonID uuid.UUID
	NewResult   string
}

type UpdateResultDraftPolicy struct {
	policy.BasePolicy
	hackathonRepo HackathonRepository
	parClient     ParticipationAndRolesClient
}

func NewUpdateResultDraftPolicy(hackathonRepo HackathonRepository, parClient ParticipationAndRolesClient) *UpdateResultDraftPolicy {
	return &UpdateResultDraftPolicy{
		BasePolicy:    policy.NewBasePolicy(ActionUpdateResultDraft),
		hackathonRepo: hackathonRepo,
		parClient:     parClient,
	}
}

func (p *UpdateResultDraftPolicy) LoadContext(ctx context.Context, params UpdateResultDraftParams) (policy.PolicyContext, error) {
	return loadHackathonPolicyContext(ctx, params.HackathonID, p.hackathonRepo, p.parClient)
}

func (p *UpdateResultDraftPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
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
			Message: "only owners and organizers can update result",
		})
		return decision
	}

	stage := domain.HackathonStage(hctx.Stage())
	if stage != domain.StageJudging {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "result can only be updated during JUDGING stage",
		})
		return decision
	}

	if hctx.ResultPublishedAt() != nil {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "result has already been published",
		})
		return decision
	}

	return decision
}

var _ policy.Policy[UpdateResultDraftParams] = (*UpdateResultDraftPolicy)(nil)

type PublishResultParams struct {
	HackathonID uuid.UUID
}

type PublishResultPolicy struct {
	policy.BasePolicy
	hackathonRepo HackathonRepository
	parClient     ParticipationAndRolesClient
}

func NewPublishResultPolicy(hackathonRepo HackathonRepository, parClient ParticipationAndRolesClient) *PublishResultPolicy {
	return &PublishResultPolicy{
		BasePolicy:    policy.NewBasePolicy(ActionPublishResult),
		hackathonRepo: hackathonRepo,
		parClient:     parClient,
	}
}

func (p *PublishResultPolicy) LoadContext(ctx context.Context, params PublishResultParams) (policy.PolicyContext, error) {
	return loadHackathonPolicyContext(ctx, params.HackathonID, p.hackathonRepo, p.parClient)
}

func (p *PublishResultPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
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
			Message: "only owners and organizers can publish result",
		})
		return decision
	}

	// Rule: stage == JUDGING && result_published_at == null
	stage := domain.HackathonStage(hctx.Stage())
	if stage != domain.StageJudging {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "result can only be published during JUDGING stage",
		})
		return decision
	}

	if hctx.ResultPublishedAt() != nil {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "result has already been published",
		})
		return decision
	}

	return decision
}

var _ policy.Policy[PublishResultParams] = (*PublishResultPolicy)(nil)
