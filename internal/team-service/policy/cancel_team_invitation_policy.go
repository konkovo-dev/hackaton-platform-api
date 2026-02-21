package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type CancelTeamInvitationParams struct {
	HackathonID  uuid.UUID
	TeamID       uuid.UUID
	InvitationID uuid.UUID
}

type CancelTeamInvitationContext struct {
	authenticated      bool
	actorUserID        uuid.UUID
	hackathonStage     string
	allowTeam          bool
	isCaptain          bool
	invitationStatus   string
	invitationBelongs  bool
}

func NewCancelTeamInvitationContext() *CancelTeamInvitationContext {
	return &CancelTeamInvitationContext{}
}

func (c *CancelTeamInvitationContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *CancelTeamInvitationContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *CancelTeamInvitationContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *CancelTeamInvitationContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *CancelTeamInvitationContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *CancelTeamInvitationContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *CancelTeamInvitationContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *CancelTeamInvitationContext) AllowTeam() bool {
	return c.allowTeam
}

func (c *CancelTeamInvitationContext) SetIsCaptain(isCaptain bool) {
	c.isCaptain = isCaptain
}

func (c *CancelTeamInvitationContext) IsCaptain() bool {
	return c.isCaptain
}

func (c *CancelTeamInvitationContext) SetInvitationStatus(status string) {
	c.invitationStatus = status
}

func (c *CancelTeamInvitationContext) InvitationStatus() string {
	return c.invitationStatus
}

func (c *CancelTeamInvitationContext) SetInvitationBelongs(belongs bool) {
	c.invitationBelongs = belongs
}

func (c *CancelTeamInvitationContext) InvitationBelongs() bool {
	return c.invitationBelongs
}

type CancelTeamInvitationPolicy struct{}

func NewCancelTeamInvitationPolicy() *CancelTeamInvitationPolicy {
	return &CancelTeamInvitationPolicy{}
}

func (p *CancelTeamInvitationPolicy) Action() policy.Action {
	return ActionCancelTeamInvitation
}

func (p *CancelTeamInvitationPolicy) LoadContext(ctx context.Context, params CancelTeamInvitationParams) (*CancelTeamInvitationContext, error) {
	return NewCancelTeamInvitationContext(), nil
}

func (p *CancelTeamInvitationPolicy) Check(ctx context.Context, pctx *CancelTeamInvitationContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if !pctx.IsCaptain() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only team captain can cancel invitations",
		})
		return decision
	}

	if pctx.HackathonStage() != "registration" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeStageRule,
			Message: "invitations can only be canceled during REGISTRATION stage",
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

	if !pctx.InvitationBelongs() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeNotFound,
			Message: "invitation not found",
		})
		return decision
	}

	if pctx.InvitationStatus() != "pending" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "only pending invitations can be canceled",
		})
		return decision
	}

	decision.Allow()
	return decision
}
