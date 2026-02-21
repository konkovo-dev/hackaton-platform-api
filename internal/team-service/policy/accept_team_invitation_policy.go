package policy

import (
	"context"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type AcceptTeamInvitationParams struct {
	InvitationID uuid.UUID
}

type AcceptTeamInvitationContext struct {
	authenticated       bool
	actorUserID         uuid.UUID
	hackathonStage      string
	allowTeam           bool
	invitationStatus    string
	invitationTarget    uuid.UUID
	isStaff             bool
	participationStatus string
}

func NewAcceptTeamInvitationContext() *AcceptTeamInvitationContext {
	return &AcceptTeamInvitationContext{}
}

func (c *AcceptTeamInvitationContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *AcceptTeamInvitationContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *AcceptTeamInvitationContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *AcceptTeamInvitationContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *AcceptTeamInvitationContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *AcceptTeamInvitationContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *AcceptTeamInvitationContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *AcceptTeamInvitationContext) AllowTeam() bool {
	return c.allowTeam
}

func (c *AcceptTeamInvitationContext) SetInvitationStatus(status string) {
	c.invitationStatus = status
}

func (c *AcceptTeamInvitationContext) InvitationStatus() string {
	return c.invitationStatus
}

func (c *AcceptTeamInvitationContext) SetInvitationTarget(target uuid.UUID) {
	c.invitationTarget = target
}

func (c *AcceptTeamInvitationContext) InvitationTarget() uuid.UUID {
	return c.invitationTarget
}

func (c *AcceptTeamInvitationContext) SetIsStaff(isStaff bool) {
	c.isStaff = isStaff
}

func (c *AcceptTeamInvitationContext) IsStaff() bool {
	return c.isStaff
}

func (c *AcceptTeamInvitationContext) SetParticipationStatus(status string) {
	c.participationStatus = status
}

func (c *AcceptTeamInvitationContext) ParticipationStatus() string {
	return c.participationStatus
}

func (c *AcceptTeamInvitationContext) IsTeamMember() bool {
	status := strings.ToLower(c.participationStatus)
	return status == "team_member" || status == "team_captain"
}

type AcceptTeamInvitationPolicy struct{}

func NewAcceptTeamInvitationPolicy() *AcceptTeamInvitationPolicy {
	return &AcceptTeamInvitationPolicy{}
}

func (p *AcceptTeamInvitationPolicy) Action() policy.Action {
	return ActionAcceptTeamInvitation
}

func (p *AcceptTeamInvitationPolicy) LoadContext(ctx context.Context, params AcceptTeamInvitationParams) (*AcceptTeamInvitationContext, error) {
	return NewAcceptTeamInvitationContext(), nil
}

func (p *AcceptTeamInvitationPolicy) Check(ctx context.Context, pctx *AcceptTeamInvitationContext) *policy.Decision {
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
			Message: "invitations can only be accepted during REGISTRATION stage",
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
			Message: "only pending invitations can be accepted",
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

	if pctx.IsStaff() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "staff members cannot join teams",
		})
		return decision
	}

	if pctx.IsTeamMember() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "you are already in a team",
		})
		return decision
	}

	decision.Allow()
	return decision
}
