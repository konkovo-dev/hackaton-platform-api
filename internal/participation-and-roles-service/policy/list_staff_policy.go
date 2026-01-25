package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ListStaffParams struct {
	HackathonID uuid.UUID
}

type StaffRoleRepository interface {
	GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error)
}

type ListStaffPolicy struct {
	policy.BasePolicy
	roleRepo StaffRoleRepository
}

func NewListStaffPolicy(roleRepo StaffRoleRepository) *ListStaffPolicy {
	return &ListStaffPolicy{
		BasePolicy: policy.NewBasePolicy(ActionReadStaff),
		roleRepo:   roleRepo,
	}
}

func (p *ListStaffPolicy) LoadContext(ctx context.Context, params ListStaffParams) (policy.PolicyContext, error) {
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

func (p *ListStaffPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
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
	hasMentorRole := sctx.HasRole(string(domain.RoleMentor))
	hasJudgeRole := sctx.HasRole(string(domain.RoleJudge))

	if !hasOwnerRole && !hasOrganizerRole && !hasMentorRole && !hasJudgeRole {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only staff members can view staff list",
		})
	}

	return decision
}

var _ policy.Policy[ListStaffParams] = (*ListStaffPolicy)(nil)
