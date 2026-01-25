package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type AnnouncementPolicyParams struct {
	HackathonID uuid.UUID
}

type CreateAnnouncementPolicy struct {
	policy.BasePolicy
	hackathonRepo HackathonRepository
	parClient     ParticipationAndRolesClient
}

func NewCreateAnnouncementPolicy(hackathonRepo HackathonRepository, parClient ParticipationAndRolesClient) *CreateAnnouncementPolicy {
	return &CreateAnnouncementPolicy{
		BasePolicy:    policy.NewBasePolicy(ActionCreateAnnouncement),
		hackathonRepo: hackathonRepo,
		parClient:     parClient,
	}
}

func (p *CreateAnnouncementPolicy) LoadContext(ctx context.Context, params AnnouncementPolicyParams) (policy.PolicyContext, error) {
	return loadHackathonPolicyContext(ctx, params.HackathonID, p.hackathonRepo, p.parClient)
}

func (p *CreateAnnouncementPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	hctx := pctx.(*HackathonPolicyContext)

	if !hctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	if hctx.Stage() == string(domain.StageDraft) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "announcements cannot be created in draft stage",
		})
		return decision
	}

	hasOwnerRole := hctx.HasRole(string(domain.RoleOwner))
	hasOrganizerRole := hctx.HasRole(string(domain.RoleOrganizer))

	if !hasOwnerRole && !hasOrganizerRole {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only owners and organizers can create announcements",
		})
	}

	return decision
}

var _ policy.Policy[AnnouncementPolicyParams] = (*CreateAnnouncementPolicy)(nil)

type ReadAnnouncementsPolicy struct {
	policy.BasePolicy
	hackathonRepo HackathonRepository
	parClient     ParticipationAndRolesClient
}

func NewReadAnnouncementsPolicy(hackathonRepo HackathonRepository, parClient ParticipationAndRolesClient) *ReadAnnouncementsPolicy {
	return &ReadAnnouncementsPolicy{
		BasePolicy:    policy.NewBasePolicy(ActionReadAnnouncements),
		hackathonRepo: hackathonRepo,
		parClient:     parClient,
	}
}

func (p *ReadAnnouncementsPolicy) LoadContext(ctx context.Context, params AnnouncementPolicyParams) (policy.PolicyContext, error) {
	return loadHackathonPolicyContext(ctx, params.HackathonID, p.hackathonRepo, p.parClient)
}

func (p *ReadAnnouncementsPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	hctx := pctx.(*HackathonPolicyContext)

	if !hctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	if hctx.Stage() == string(domain.StageDraft) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "announcements cannot be read in draft stage",
		})
		return decision
	}

	isStaff := hctx.HasRole(string(domain.RoleOwner)) ||
		hctx.HasRole(string(domain.RoleOrganizer)) ||
		hctx.HasRole(string(domain.RoleMentor)) ||
		hctx.HasRole(string(domain.RoleJury))

	isParticipant := hctx.ParticipationKind() == string(domain.ParticipationLookingForTeam) ||
		hctx.ParticipationKind() == string(domain.ParticipationIndividual) ||
		hctx.ParticipationKind() == string(domain.ParticipationTeamMember) ||
		hctx.ParticipationKind() == string(domain.ParticipationTeamCaptain)

	if !isStaff && !isParticipant {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only staff or participants can read announcements",
		})
	}

	return decision
}

var _ policy.Policy[AnnouncementPolicyParams] = (*ReadAnnouncementsPolicy)(nil)

type UpdateAnnouncementPolicy struct {
	policy.BasePolicy
	hackathonRepo HackathonRepository
	parClient     ParticipationAndRolesClient
}

func NewUpdateAnnouncementPolicy(hackathonRepo HackathonRepository, parClient ParticipationAndRolesClient) *UpdateAnnouncementPolicy {
	return &UpdateAnnouncementPolicy{
		BasePolicy:    policy.NewBasePolicy(ActionUpdateAnnouncement),
		hackathonRepo: hackathonRepo,
		parClient:     parClient,
	}
}

func (p *UpdateAnnouncementPolicy) LoadContext(ctx context.Context, params AnnouncementPolicyParams) (policy.PolicyContext, error) {
	return loadHackathonPolicyContext(ctx, params.HackathonID, p.hackathonRepo, p.parClient)
}

func (p *UpdateAnnouncementPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	hctx := pctx.(*HackathonPolicyContext)

	if !hctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	if hctx.Stage() == string(domain.StageDraft) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "announcements cannot be updated in draft stage",
		})
		return decision
	}

	hasOwnerRole := hctx.HasRole(string(domain.RoleOwner))
	hasOrganizerRole := hctx.HasRole(string(domain.RoleOrganizer))

	if !hasOwnerRole && !hasOrganizerRole {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only owners and organizers can update announcements",
		})
	}

	return decision
}

var _ policy.Policy[AnnouncementPolicyParams] = (*UpdateAnnouncementPolicy)(nil)

type DeleteAnnouncementPolicy struct {
	policy.BasePolicy
	hackathonRepo HackathonRepository
	parClient     ParticipationAndRolesClient
}

func NewDeleteAnnouncementPolicy(hackathonRepo HackathonRepository, parClient ParticipationAndRolesClient) *DeleteAnnouncementPolicy {
	return &DeleteAnnouncementPolicy{
		BasePolicy:    policy.NewBasePolicy(ActionDeleteAnnouncement),
		hackathonRepo: hackathonRepo,
		parClient:     parClient,
	}
}

func (p *DeleteAnnouncementPolicy) LoadContext(ctx context.Context, params AnnouncementPolicyParams) (policy.PolicyContext, error) {
	return loadHackathonPolicyContext(ctx, params.HackathonID, p.hackathonRepo, p.parClient)
}

func (p *DeleteAnnouncementPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	hctx := pctx.(*HackathonPolicyContext)

	if !hctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	if hctx.Stage() == string(domain.StageDraft) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "announcements cannot be deleted in draft stage",
		})
		return decision
	}

	hasOwnerRole := hctx.HasRole(string(domain.RoleOwner))
	hasOrganizerRole := hctx.HasRole(string(domain.RoleOrganizer))

	if !hasOwnerRole && !hasOrganizerRole {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only owners and organizers can delete announcements",
		})
	}

	return decision
}

var _ policy.Policy[AnnouncementPolicyParams] = (*DeleteAnnouncementPolicy)(nil)
