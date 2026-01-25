package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type RemoveHackathonRoleParams struct {
	HackathonID  uuid.UUID
	TargetUserID uuid.UUID
	Role         string
}

type RemoveHackathonRolePolicyContext struct {
	StaffPolicyContext
	targetUserID      uuid.UUID
	roleToRemove      string
	targetUserHasRole bool
}

func NewRemoveHackathonRolePolicyContext() *RemoveHackathonRolePolicyContext {
	return &RemoveHackathonRolePolicyContext{}
}

func (c *RemoveHackathonRolePolicyContext) SetTargetUserID(userID uuid.UUID) {
	c.targetUserID = userID
}

func (c *RemoveHackathonRolePolicyContext) TargetUserID() uuid.UUID {
	return c.targetUserID
}

func (c *RemoveHackathonRolePolicyContext) SetRoleToRemove(role string) {
	c.roleToRemove = role
}

func (c *RemoveHackathonRolePolicyContext) RoleToRemove() string {
	return c.roleToRemove
}

func (c *RemoveHackathonRolePolicyContext) SetTargetUserHasRole(hasRole bool) {
	c.targetUserHasRole = hasRole
}

func (c *RemoveHackathonRolePolicyContext) TargetUserHasRole() bool {
	return c.targetUserHasRole
}

var _ policy.PolicyContext = (*RemoveHackathonRolePolicyContext)(nil)

type RemovePolicyRepository interface {
	GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error)
	HasRole(ctx context.Context, hackathonID, userID uuid.UUID, role string) (bool, error)
}

type RemoveHackathonRolePolicy struct {
	policy.BasePolicy
	roleRepo RemovePolicyRepository
}

func NewRemoveHackathonRolePolicy(roleRepo RemovePolicyRepository) *RemoveHackathonRolePolicy {
	return &RemoveHackathonRolePolicy{
		BasePolicy: policy.NewBasePolicy(ActionRemoveRole),
		roleRepo:   roleRepo,
	}
}

func (p *RemoveHackathonRolePolicy) LoadContext(ctx context.Context, params RemoveHackathonRoleParams) (policy.PolicyContext, error) {
	pctx := NewRemoveHackathonRolePolicyContext()
	pctx.SetTargetUserID(params.TargetUserID)
	pctx.SetRoleToRemove(params.Role)

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

	actorRoles, err := p.roleRepo.GetRoleStrings(ctx, params.HackathonID, userUUID)
	if err == nil {
		pctx.SetRoles(actorRoles)
	}

	targetHasRole, err := p.roleRepo.HasRole(ctx, params.HackathonID, params.TargetUserID, params.Role)
	if err == nil {
		pctx.SetTargetUserHasRole(targetHasRole)
	}

	return pctx, nil
}

func (p *RemoveHackathonRolePolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	rCtx := pctx.(*RemoveHackathonRolePolicyContext)

	if !rCtx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	if !rCtx.HasRole(string(domain.RoleOwner)) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only OWNER can remove roles",
		})
		return decision
	}

	if rCtx.RoleToRemove() == string(domain.RoleOwner) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "cannot remove OWNER role",
		})
		return decision
	}

	if !rCtx.TargetUserHasRole() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeNotFound,
			Message: "target user does not have this role",
		})
		return decision
	}

	return decision
}

var _ policy.Policy[RemoveHackathonRoleParams] = (*RemoveHackathonRolePolicy)(nil)
