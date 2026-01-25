package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type CancelStaffInvitationParams struct {
	HackathonID  uuid.UUID
	InvitationID uuid.UUID
}

type CancelInvitationContext struct {
	authenticated bool
	actorUserID   uuid.UUID

	actorRoles []string

	invitationID     uuid.UUID
	invitationStatus string
	hackathonID      uuid.UUID
}

func NewCancelInvitationContext() *CancelInvitationContext {
	return &CancelInvitationContext{}
}

func (c *CancelInvitationContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *CancelInvitationContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *CancelInvitationContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *CancelInvitationContext) SetActorUserID(id uuid.UUID) {
	c.actorUserID = id
}

func (c *CancelInvitationContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *CancelInvitationContext) ActorHasRole(role string) bool {
	for _, r := range c.actorRoles {
		if r == role {
			return true
		}
	}
	return false
}

func (c *CancelInvitationContext) SetInvitationID(id uuid.UUID) {
	c.invitationID = id
}

func (c *CancelInvitationContext) SetInvitationStatus(status string) {
	c.invitationStatus = status
}

func (c *CancelInvitationContext) InvitationStatus() string {
	return c.invitationStatus
}

func (c *CancelInvitationContext) SetHackathonID(id uuid.UUID) {
	c.hackathonID = id
}

func (c *CancelInvitationContext) HackathonID() uuid.UUID {
	return c.hackathonID
}

type CancelStaffInvitationRepositories interface {
	GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error)
	GetInvitationStatus(ctx context.Context, invitationID uuid.UUID) (status string, hackathonID uuid.UUID, err error)
}

type CancelStaffInvitationPolicy struct {
	policy.BasePolicy
	repos CancelStaffInvitationRepositories
}

func NewCancelStaffInvitationPolicy(repos CancelStaffInvitationRepositories) *CancelStaffInvitationPolicy {
	return &CancelStaffInvitationPolicy{
		BasePolicy: policy.NewBasePolicy(ActionCancelInvitation),
		repos:      repos,
	}
}

func (p *CancelStaffInvitationPolicy) LoadContext(ctx context.Context, params CancelStaffInvitationParams) (policy.PolicyContext, error) {
	pctx := NewCancelInvitationContext()
	pctx.SetInvitationID(params.InvitationID)
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

	invitationStatus, invitationHackathonID, err := p.repos.GetInvitationStatus(ctx, params.InvitationID)
	if err == nil {
		pctx.SetInvitationStatus(invitationStatus)
		if invitationHackathonID != params.HackathonID {
			pctx.SetInvitationStatus("")
		}
	}

	actorRoles, err := p.repos.GetRoleStrings(ctx, params.HackathonID, userUUID)
	if err == nil {
		pctx.SetActorRoles(actorRoles)
	}

	return pctx, nil
}

func (p *CancelStaffInvitationPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	cancelCtx := pctx.(*CancelInvitationContext)

	if !cancelCtx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	if !cancelCtx.ActorHasRole(string(domain.RoleOwner)) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only OWNER can cancel staff invitations",
		})
		return decision
	}

	if cancelCtx.InvitationStatus() == "" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeNotFound,
			Message: "invitation not found",
		})
		return decision
	}

	if cancelCtx.InvitationStatus() != string(domain.InvitationStatusPending) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "can only cancel pending invitations",
		})
		return decision
	}

	return decision
}

var _ policy.Policy[CancelStaffInvitationParams] = (*CancelStaffInvitationPolicy)(nil)
