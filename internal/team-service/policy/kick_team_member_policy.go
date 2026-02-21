package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type KickTeamMemberParams struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
	TargetUserID uuid.UUID
}

type KickTeamMemberContext struct {
	authenticated    bool
	actorUserID      uuid.UUID
	hackathonStage   string
	allowTeam        bool
	isCaptain        bool
	targetIsMember   bool
	targetIsCaptain  bool
	targetUserID     uuid.UUID
}

func NewKickTeamMemberContext() *KickTeamMemberContext {
	return &KickTeamMemberContext{}
}

func (c *KickTeamMemberContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *KickTeamMemberContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *KickTeamMemberContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *KickTeamMemberContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *KickTeamMemberContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *KickTeamMemberContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *KickTeamMemberContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *KickTeamMemberContext) AllowTeam() bool {
	return c.allowTeam
}

func (c *KickTeamMemberContext) SetIsCaptain(isCaptain bool) {
	c.isCaptain = isCaptain
}

func (c *KickTeamMemberContext) IsCaptain() bool {
	return c.isCaptain
}

func (c *KickTeamMemberContext) SetTargetIsMember(isMember bool) {
	c.targetIsMember = isMember
}

func (c *KickTeamMemberContext) TargetIsMember() bool {
	return c.targetIsMember
}

func (c *KickTeamMemberContext) SetTargetIsCaptain(isCaptain bool) {
	c.targetIsCaptain = isCaptain
}

func (c *KickTeamMemberContext) TargetIsCaptain() bool {
	return c.targetIsCaptain
}

func (c *KickTeamMemberContext) SetTargetUserID(userID uuid.UUID) {
	c.targetUserID = userID
}

func (c *KickTeamMemberContext) TargetUserID() uuid.UUID {
	return c.targetUserID
}

type KickTeamMemberPolicy struct{}

func NewKickTeamMemberPolicy() *KickTeamMemberPolicy {
	return &KickTeamMemberPolicy{}
}

func (p *KickTeamMemberPolicy) Action() policy.Action {
	return ActionKickTeamMember
}

func (p *KickTeamMemberPolicy) LoadContext(ctx context.Context, params KickTeamMemberParams) (*KickTeamMemberContext, error) {
	return NewKickTeamMemberContext(), nil
}

func (p *KickTeamMemberPolicy) Check(ctx context.Context, pctx *KickTeamMemberContext) *policy.Decision {
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
			Message: "only team captain can kick members",
		})
		return decision
	}

	if pctx.HackathonStage() != "registration" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeStageRule,
			Message: "members can only be kicked during REGISTRATION stage",
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

	if !pctx.TargetIsMember() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeNotFound,
			Message: "target user is not a team member",
		})
		return decision
	}

	if pctx.TargetIsCaptain() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "cannot kick the team captain",
		})
		return decision
	}

	if pctx.TargetUserID() == pctx.ActorUserID() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "cannot kick yourself",
		})
		return decision
	}

	decision.Allow()
	return decision
}
