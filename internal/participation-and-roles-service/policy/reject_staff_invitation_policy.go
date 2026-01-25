package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type RejectStaffInvitationParams struct {
	InvitationID uuid.UUID
}

type RejectInvitationContext struct {
	authenticated bool
	actorUserID   uuid.UUID

	invitationExists     bool
	invitationStatus     string
	invitationTargetUser uuid.UUID
}

func NewRejectInvitationContext() *RejectInvitationContext {
	return &RejectInvitationContext{}
}

func (c *RejectInvitationContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *RejectInvitationContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *RejectInvitationContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *RejectInvitationContext) SetActorUserID(id uuid.UUID) {
	c.actorUserID = id
}

func (c *RejectInvitationContext) SetInvitationExists(exists bool) {
	c.invitationExists = exists
}

func (c *RejectInvitationContext) InvitationExists() bool {
	return c.invitationExists
}

func (c *RejectInvitationContext) SetInvitationStatus(status string) {
	c.invitationStatus = status
}

func (c *RejectInvitationContext) InvitationStatus() string {
	return c.invitationStatus
}

func (c *RejectInvitationContext) SetInvitationTargetUser(userID uuid.UUID) {
	c.invitationTargetUser = userID
}

func (c *RejectInvitationContext) InvitationTargetUser() uuid.UUID {
	return c.invitationTargetUser
}

type RejectStaffInvitationRepositories interface {
	GetInvitationBasicInfo(ctx context.Context, invitationID uuid.UUID) (exists bool, status string, targetUserID uuid.UUID, err error)
}

type RejectStaffInvitationPolicy struct {
	policy.BasePolicy
	repos RejectStaffInvitationRepositories
}

func NewRejectStaffInvitationPolicy(repos RejectStaffInvitationRepositories) *RejectStaffInvitationPolicy {
	return &RejectStaffInvitationPolicy{
		BasePolicy: policy.NewBasePolicy(ActionRejectInvitation),
		repos:      repos,
	}
}

func (p *RejectStaffInvitationPolicy) LoadContext(ctx context.Context, params RejectStaffInvitationParams) (policy.PolicyContext, error) {
	pctx := NewRejectInvitationContext()

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

	exists, status, targetUserID, err := p.repos.GetInvitationBasicInfo(ctx, params.InvitationID)
	if err == nil && exists {
		pctx.SetInvitationExists(true)
		pctx.SetInvitationStatus(status)
		pctx.SetInvitationTargetUser(targetUserID)
	}

	return pctx, nil
}

func (p *RejectStaffInvitationPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	rejectCtx := pctx.(*RejectInvitationContext)

	if !rejectCtx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	if !rejectCtx.InvitationExists() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeNotFound,
			Message: "invitation not found",
		})
		return decision
	}

	if rejectCtx.InvitationStatus() != string(domain.InvitationStatusPending) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "can only reject pending invitations",
		})
		return decision
	}

	if rejectCtx.InvitationTargetUser() != rejectCtx.ActorUserID() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "can only reject invitations addressed to you",
		})
		return decision
	}

	return decision
}

var _ policy.Policy[RejectStaffInvitationParams] = (*RejectStaffInvitationPolicy)(nil)
