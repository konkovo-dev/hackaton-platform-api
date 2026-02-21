package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type RejectTeamInvitationParams struct {
	InvitationID uuid.UUID
}

type RejectTeamInvitationContext struct {
	authenticated    bool
	actorUserID      uuid.UUID
	hackathonStage   string
	allowTeam        bool
	invitationStatus string
	invitationTarget uuid.UUID
}

func NewRejectTeamInvitationContext() *RejectTeamInvitationContext {
	return &RejectTeamInvitationContext{}
}

func (c *RejectTeamInvitationContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *RejectTeamInvitationContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *RejectTeamInvitationContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *RejectTeamInvitationContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *RejectTeamInvitationContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *RejectTeamInvitationContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *RejectTeamInvitationContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *RejectTeamInvitationContext) AllowTeam() bool {
	return c.allowTeam
}

func (c *RejectTeamInvitationContext) SetInvitationStatus(status string) {
	c.invitationStatus = status
}

func (c *RejectTeamInvitationContext) InvitationStatus() string {
	return c.invitationStatus
}

func (c *RejectTeamInvitationContext) SetInvitationTarget(target uuid.UUID) {
	c.invitationTarget = target
}

func (c *RejectTeamInvitationContext) InvitationTarget() uuid.UUID {
	return c.invitationTarget
}

type RejectTeamInvitationPolicy struct{}

func NewRejectTeamInvitationPolicy() *RejectTeamInvitationPolicy {
	return &RejectTeamInvitationPolicy{}
}

func (p *RejectTeamInvitationPolicy) Action() policy.Action {
	return ActionRejectTeamInvitation
}

func (p *RejectTeamInvitationPolicy) LoadContext(ctx context.Context, params RejectTeamInvitationParams) (*RejectTeamInvitationContext, error) {
	return NewRejectTeamInvitationContext(), nil
}

func (p *RejectTeamInvitationPolicy) Check(ctx context.Context, pctx *RejectTeamInvitationContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if pctx.HackathonStage() != "registration" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeStageRule,
			Message: "invitations can only be rejected during REGISTRATION stage",
		})
		return decision
	}

	if !pctx.AllowTeam() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "team operations are not allowed for this hackathon",
		})
		return decision
	}

	if pctx.InvitationStatus() != "pending" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "only pending invitations can be rejected",
		})
		return decision
	}

	if pctx.InvitationTarget() != pctx.ActorUserID() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "invitation is not addressed to you",
		})
		return decision
	}

	decision.Allow()
	return decision
}
