package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type SelfRemoveHackathonRoleParams struct {
	HackathonID uuid.UUID
	Role        string
}

type SelfRemoveHackathonRolePolicyContext struct {
	StaffPolicyContext
	roleToRemove string
	actorHasRole bool
}

func NewSelfRemoveHackathonRolePolicyContext() *SelfRemoveHackathonRolePolicyContext {
	return &SelfRemoveHackathonRolePolicyContext{}
}

func (c *SelfRemoveHackathonRolePolicyContext) SetRoleToRemove(role string) {
	c.roleToRemove = role
}

func (c *SelfRemoveHackathonRolePolicyContext) RoleToRemove() string {
	return c.roleToRemove
}

func (c *SelfRemoveHackathonRolePolicyContext) SetActorHasRole(hasRole bool) {
	c.actorHasRole = hasRole
}

func (c *SelfRemoveHackathonRolePolicyContext) ActorHasRole() bool {
	return c.actorHasRole
}

var _ policy.PolicyContext = (*SelfRemoveHackathonRolePolicyContext)(nil)

type SelfRemovePolicyRepository interface {
	HasRole(ctx context.Context, hackathonID, userID uuid.UUID, role string) (bool, error)
}

type SelfRemoveHackathonRolePolicy struct {
	policy.BasePolicy
	roleRepo SelfRemovePolicyRepository
}

func NewSelfRemoveHackathonRolePolicy(roleRepo SelfRemovePolicyRepository) *SelfRemoveHackathonRolePolicy {
	return &SelfRemoveHackathonRolePolicy{
		BasePolicy: policy.NewBasePolicy(ActionSelfRemoveRole),
		roleRepo:   roleRepo,
	}
}

func (p *SelfRemoveHackathonRolePolicy) LoadContext(ctx context.Context, params SelfRemoveHackathonRoleParams) (policy.PolicyContext, error) {
	pctx := NewSelfRemoveHackathonRolePolicyContext()
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

	actorHasRole, err := p.roleRepo.HasRole(ctx, params.HackathonID, userUUID, params.Role)
	if err == nil {
		pctx.SetActorHasRole(actorHasRole)
	}

	return pctx, nil
}

func (p *SelfRemoveHackathonRolePolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	srCtx := pctx.(*SelfRemoveHackathonRolePolicyContext)

	if !srCtx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	if srCtx.RoleToRemove() == string(domain.RoleOwner) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "cannot remove OWNER role",
		})
		return decision
	}

	if !srCtx.ActorHasRole() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeNotFound,
			Message: "you do not have this role",
		})
		return decision
	}

	return decision
}

var _ policy.Policy[SelfRemoveHackathonRoleParams] = (*SelfRemoveHackathonRolePolicy)(nil)
