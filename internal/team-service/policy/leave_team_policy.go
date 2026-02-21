package policy

import (
	"context"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type LeaveTeamParams struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
}

type LeaveTeamContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	allowTeam      bool
	isMember       bool
	isCaptain      bool
}

func NewLeaveTeamContext() *LeaveTeamContext {
	return &LeaveTeamContext{}
}

func (c *LeaveTeamContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *LeaveTeamContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *LeaveTeamContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *LeaveTeamContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *LeaveTeamContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *LeaveTeamContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *LeaveTeamContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *LeaveTeamContext) AllowTeam() bool {
	return c.allowTeam
}

func (c *LeaveTeamContext) SetIsMember(isMember bool) {
	c.isMember = isMember
}

func (c *LeaveTeamContext) IsMember() bool {
	return c.isMember
}

func (c *LeaveTeamContext) SetIsCaptain(isCaptain bool) {
	c.isCaptain = isCaptain
}

func (c *LeaveTeamContext) IsCaptain() bool {
	return c.isCaptain
}

func (c *LeaveTeamContext) IsInTeamLeaveWindow() bool {
	stage := strings.ToLower(c.hackathonStage)
	return stage == "registration" || stage == "prestart" || stage == "running" || stage == "judging" || stage == "finished"
}

type LeaveTeamPolicy struct{}

func NewLeaveTeamPolicy() *LeaveTeamPolicy {
	return &LeaveTeamPolicy{}
}

func (p *LeaveTeamPolicy) Action() policy.Action {
	return ActionLeaveTeam
}

func (p *LeaveTeamPolicy) LoadContext(ctx context.Context, params LeaveTeamParams) (*LeaveTeamContext, error) {
	return NewLeaveTeamContext(), nil
}

func (p *LeaveTeamPolicy) Check(ctx context.Context, pctx *LeaveTeamContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if !pctx.IsInTeamLeaveWindow() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeStageRule,
			Message: "team can only be left during REGISTRATION, PRESTART, RUNNING, JUDGING, or FINISHED stages",
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

	if !pctx.IsMember() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeNotFound,
			Message: "you are not a member of this team",
		})
		return decision
	}

	if pctx.IsCaptain() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "captain cannot leave the team",
		})
		return decision
	}

	decision.Allow()
	return decision
}
