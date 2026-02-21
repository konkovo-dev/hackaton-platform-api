package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type CreateTeamInvitationParams struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
}

type CreateTeamInvitationContext struct {
	authenticated        bool
	actorUserID          uuid.UUID
	hackathonStage       string
	allowTeam            bool
	isCaptain            bool
	targetIsStaff        bool
	isTargetTeamMember   bool
}

func NewCreateTeamInvitationContext() *CreateTeamInvitationContext {
	return &CreateTeamInvitationContext{}
}

func (c *CreateTeamInvitationContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *CreateTeamInvitationContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *CreateTeamInvitationContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *CreateTeamInvitationContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *CreateTeamInvitationContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *CreateTeamInvitationContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *CreateTeamInvitationContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *CreateTeamInvitationContext) AllowTeam() bool {
	return c.allowTeam
}

func (c *CreateTeamInvitationContext) SetIsCaptain(isCaptain bool) {
	c.isCaptain = isCaptain
}

func (c *CreateTeamInvitationContext) IsCaptain() bool {
	return c.isCaptain
}

func (c *CreateTeamInvitationContext) SetTargetIsStaff(isStaff bool) {
	c.targetIsStaff = isStaff
}

func (c *CreateTeamInvitationContext) TargetIsStaff() bool {
	return c.targetIsStaff
}

func (c *CreateTeamInvitationContext) SetIsTargetTeamMember(isTeamMember bool) {
	c.isTargetTeamMember = isTeamMember
}

func (c *CreateTeamInvitationContext) IsTargetTeamMember() bool {
	return c.isTargetTeamMember
}

type CreateTeamInvitationPolicy struct{}

func NewCreateTeamInvitationPolicy() *CreateTeamInvitationPolicy {
	return &CreateTeamInvitationPolicy{}
}

func (p *CreateTeamInvitationPolicy) Action() policy.Action {
	return ActionCreateTeamInvitation
}

func (p *CreateTeamInvitationPolicy) LoadContext(ctx context.Context, params CreateTeamInvitationParams) (*CreateTeamInvitationContext, error) {
	return NewCreateTeamInvitationContext(), nil
}

func (p *CreateTeamInvitationPolicy) Check(ctx context.Context, pctx *CreateTeamInvitationContext) *policy.Decision {
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
			Message: "team invitations can only be created during REGISTRATION stage",
		})
		return decision
	}

	if !pctx.AllowTeam() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "team invitations are not allowed for this hackathon",
		})
		return decision
	}

	if !pctx.IsCaptain() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only team captain can create invitations",
		})
		return decision
	}

	if pctx.TargetIsStaff() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "cannot invite staff members",
		})
		return decision
	}

	if pctx.IsTargetTeamMember() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "target user is already in a team",
		})
		return decision
	}

	decision.Allow()
	return decision
}
