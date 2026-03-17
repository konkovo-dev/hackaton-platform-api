package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

const ActionListHackathonStaffInvitations policy.Action = "list_hackathon_staff_invitations"

type ListHackathonStaffInvitationsParams struct {
	HackathonID uuid.UUID
}

type ListHackathonStaffInvitationsPolicy struct {
	policy.BasePolicy
	roleRepo StaffRoleRepository
}

func NewListHackathonStaffInvitationsPolicy(roleRepo StaffRoleRepository) *ListHackathonStaffInvitationsPolicy {
	return &ListHackathonStaffInvitationsPolicy{
		BasePolicy: policy.NewBasePolicy(ActionListHackathonStaffInvitations),
		roleRepo:   roleRepo,
	}
}

func (p *ListHackathonStaffInvitationsPolicy) LoadContext(ctx context.Context, params ListHackathonStaffInvitationsParams) (policy.PolicyContext, error) {
	pctx := NewStaffPolicyContext()
	pctx.SetHackathonID(params.HackathonID)

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

	roles, err := p.roleRepo.GetRoleStrings(ctx, params.HackathonID, userUUID)
	if err == nil {
		pctx.SetRoles(roles)
	}

	return pctx, nil
}

func (p *ListHackathonStaffInvitationsPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	sctx := pctx.(*StaffPolicyContext)

	if !sctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	hasOwnerRole := sctx.HasRole(string(domain.RoleOwner))
	hasOrganizerRole := sctx.HasRole(string(domain.RoleOrganizer))

	if !hasOwnerRole && !hasOrganizerRole {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only OWNER or ORGANIZER can view staff invitations",
		})
	}

	return decision
}

var _ policy.Policy[ListHackathonStaffInvitationsParams] = (*ListHackathonStaffInvitationsPolicy)(nil)
