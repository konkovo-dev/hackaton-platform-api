package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type CreateStaffInvitationParams struct {
	HackathonID   uuid.UUID
	TargetUserID  uuid.UUID
	RequestedRole string
}

type StaffInvitationContext struct {
	authenticated bool
	actorUserID   uuid.UUID

	actorRoles []string

	targetUserID            uuid.UUID
	targetRoles             []string
	targetParticipationKind string
	requestedRole           string
}

func NewStaffInvitationContext() *StaffInvitationContext {
	return &StaffInvitationContext{}
}

func (c *StaffInvitationContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *StaffInvitationContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *StaffInvitationContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *StaffInvitationContext) SetActorUserID(id uuid.UUID) {
	c.actorUserID = id
}

func (c *StaffInvitationContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *StaffInvitationContext) ActorHasRole(role string) bool {
	for _, r := range c.actorRoles {
		if r == role {
			return true
		}
	}
	return false
}

func (c *StaffInvitationContext) SetTargetUserID(id uuid.UUID) {
	c.targetUserID = id
}

func (c *StaffInvitationContext) SetTargetRoles(roles []string) {
	c.targetRoles = roles
}

func (c *StaffInvitationContext) TargetHasRole(role string) bool {
	for _, r := range c.targetRoles {
		if r == role {
			return true
		}
	}
	return false
}

func (c *StaffInvitationContext) SetTargetParticipationKind(kind string) {
	c.targetParticipationKind = kind
}

func (c *StaffInvitationContext) TargetParticipationKind() string {
	return c.targetParticipationKind
}

func (c *StaffInvitationContext) SetRequestedRole(role string) {
	c.requestedRole = role
}

func (c *StaffInvitationContext) RequestedRole() string {
	return c.requestedRole
}

type CreateStaffInvitationRepositories interface {
	GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error)
	GetParticipationStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error)
}

type CreateStaffInvitationPolicy struct {
	policy.BasePolicy
	repos CreateStaffInvitationRepositories
}

func NewCreateStaffInvitationPolicy(repos CreateStaffInvitationRepositories) *CreateStaffInvitationPolicy {
	return &CreateStaffInvitationPolicy{
		BasePolicy: policy.NewBasePolicy(ActionCreateInvitation),
		repos:      repos,
	}
}

func (p *CreateStaffInvitationPolicy) LoadContext(ctx context.Context, params CreateStaffInvitationParams) (policy.PolicyContext, error) {
	pctx := NewStaffInvitationContext()
	pctx.SetTargetUserID(params.TargetUserID)
	pctx.SetRequestedRole(params.RequestedRole)

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

	actorRoles, err := p.repos.GetRoleStrings(ctx, params.HackathonID, userUUID)
	if err == nil {
		pctx.SetActorRoles(actorRoles)
	}

	targetRoles, err := p.repos.GetRoleStrings(ctx, params.HackathonID, params.TargetUserID)
	if err == nil {
		pctx.SetTargetRoles(targetRoles)
	}

	targetParticipationStatus, err := p.repos.GetParticipationStatus(ctx, params.HackathonID, params.TargetUserID)
	if err == nil {
		pctx.SetTargetParticipationKind(targetParticipationStatus)
	}

	return pctx, nil
}

func (p *CreateStaffInvitationPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	invCtx := pctx.(*StaffInvitationContext)

	if !invCtx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	if !invCtx.ActorHasRole(string(domain.RoleOwner)) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only OWNER can create staff invitations",
		})
		return decision
	}

	requestedRole := invCtx.RequestedRole()
	if requestedRole != string(domain.RoleOrganizer) &&
		requestedRole != string(domain.RoleMentor) &&
		requestedRole != string(domain.RoleJudge) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "can only invite to ORGANIZER, MENTOR, or JUDGE roles",
		})
		return decision
	}

	if invCtx.TargetHasRole(string(domain.RoleOwner)) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "cannot invite user who is already OWNER",
		})
		return decision
	}

	if invCtx.TargetHasRole(requestedRole) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "user already has the requested role",
		})
		return decision
	}

	targetParticipationKind := invCtx.TargetParticipationKind()
	if targetParticipationKind != "" && targetParticipationKind != string(domain.ParticipationNone) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "cannot invite user who is a participant",
		})
		return decision
	}

	return decision
}

var _ policy.Policy[CreateStaffInvitationParams] = (*CreateStaffInvitationPolicy)(nil)
