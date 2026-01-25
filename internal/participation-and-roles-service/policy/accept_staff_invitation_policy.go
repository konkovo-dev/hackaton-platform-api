package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type AcceptStaffInvitationParams struct {
	InvitationID uuid.UUID
}

type AcceptInvitationContext struct {
	authenticated bool
	actorUserID   uuid.UUID

	invitationExists     bool
	invitationStatus     string
	invitationTargetUser uuid.UUID
	invitationHackathon  uuid.UUID
	requestedRole        string

	actorRoles             []string
	actorParticipationKind string
}

func NewAcceptInvitationContext() *AcceptInvitationContext {
	return &AcceptInvitationContext{}
}

func (c *AcceptInvitationContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *AcceptInvitationContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *AcceptInvitationContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *AcceptInvitationContext) SetActorUserID(id uuid.UUID) {
	c.actorUserID = id
}

func (c *AcceptInvitationContext) SetInvitationExists(exists bool) {
	c.invitationExists = exists
}

func (c *AcceptInvitationContext) InvitationExists() bool {
	return c.invitationExists
}

func (c *AcceptInvitationContext) SetInvitationStatus(status string) {
	c.invitationStatus = status
}

func (c *AcceptInvitationContext) InvitationStatus() string {
	return c.invitationStatus
}

func (c *AcceptInvitationContext) SetInvitationTargetUser(userID uuid.UUID) {
	c.invitationTargetUser = userID
}

func (c *AcceptInvitationContext) InvitationTargetUser() uuid.UUID {
	return c.invitationTargetUser
}

func (c *AcceptInvitationContext) SetInvitationHackathon(hackathonID uuid.UUID) {
	c.invitationHackathon = hackathonID
}

func (c *AcceptInvitationContext) InvitationHackathon() uuid.UUID {
	return c.invitationHackathon
}

func (c *AcceptInvitationContext) SetRequestedRole(role string) {
	c.requestedRole = role
}

func (c *AcceptInvitationContext) RequestedRole() string {
	return c.requestedRole
}

func (c *AcceptInvitationContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *AcceptInvitationContext) ActorHasRole(role string) bool {
	for _, r := range c.actorRoles {
		if r == role {
			return true
		}
	}
	return false
}

func (c *AcceptInvitationContext) SetActorParticipationKind(kind string) {
	c.actorParticipationKind = kind
}

func (c *AcceptInvitationContext) ActorParticipationKind() string {
	return c.actorParticipationKind
}

type AcceptStaffInvitationRepositories interface {
	GetInvitationDetails(ctx context.Context, invitationID uuid.UUID) (exists bool, status string, targetUserID uuid.UUID, hackathonID uuid.UUID, requestedRole string, err error)
	GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error)
	GetParticipationStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error)
}

type AcceptStaffInvitationPolicy struct {
	policy.BasePolicy
	repos AcceptStaffInvitationRepositories
}

func NewAcceptStaffInvitationPolicy(repos AcceptStaffInvitationRepositories) *AcceptStaffInvitationPolicy {
	return &AcceptStaffInvitationPolicy{
		BasePolicy: policy.NewBasePolicy(ActionAcceptInvitation),
		repos:      repos,
	}
}

func (p *AcceptStaffInvitationPolicy) LoadContext(ctx context.Context, params AcceptStaffInvitationParams) (policy.PolicyContext, error) {
	pctx := NewAcceptInvitationContext()

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

	exists, status, targetUserID, hackathonID, requestedRole, err := p.repos.GetInvitationDetails(ctx, params.InvitationID)
	if err == nil && exists {
		pctx.SetInvitationExists(true)
		pctx.SetInvitationStatus(status)
		pctx.SetInvitationTargetUser(targetUserID)
		pctx.SetInvitationHackathon(hackathonID)
		pctx.SetRequestedRole(requestedRole)

		actorRoles, err := p.repos.GetRoleStrings(ctx, hackathonID, userUUID)
		if err == nil {
			pctx.SetActorRoles(actorRoles)
		}

		actorParticipationStatus, err := p.repos.GetParticipationStatus(ctx, hackathonID, userUUID)
		if err == nil {
			pctx.SetActorParticipationKind(actorParticipationStatus)
		}
	}

	return pctx, nil
}

func (p *AcceptStaffInvitationPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	acceptCtx := pctx.(*AcceptInvitationContext)

	if !acceptCtx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	if !acceptCtx.InvitationExists() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeNotFound,
			Message: "invitation not found",
		})
		return decision
	}

	if acceptCtx.InvitationStatus() != string(domain.InvitationStatusPending) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "can only accept pending invitations",
		})
		return decision
	}

	if acceptCtx.InvitationTargetUser() != acceptCtx.ActorUserID() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "can only accept invitations addressed to you",
		})
		return decision
	}

	requestedRole := acceptCtx.RequestedRole()
	if requestedRole != string(domain.RoleOrganizer) &&
		requestedRole != string(domain.RoleMentor) &&
		requestedRole != string(domain.RoleJudge) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "invalid requested role",
		})
		return decision
	}

	if acceptCtx.ActorHasRole(string(domain.RoleOwner)) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "cannot accept invitation if already OWNER",
		})
		return decision
	}

	if acceptCtx.ActorHasRole(requestedRole) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "already has the requested role",
		})
		return decision
	}

	participationKind := acceptCtx.ActorParticipationKind()
	if participationKind != "" && participationKind != string(domain.ParticipationNone) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "cannot accept invitation if already a participant",
		})
		return decision
	}

	return decision
}

var _ policy.Policy[AcceptStaffInvitationParams] = (*AcceptStaffInvitationPolicy)(nil)
