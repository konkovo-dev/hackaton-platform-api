package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type DeleteTeamParams struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
}

type DeleteTeamContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	allowTeam      bool
	isCaptain      bool
	membersCount   int64
}

func NewDeleteTeamContext() *DeleteTeamContext {
	return &DeleteTeamContext{}
}

func (c *DeleteTeamContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *DeleteTeamContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *DeleteTeamContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *DeleteTeamContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *DeleteTeamContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *DeleteTeamContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *DeleteTeamContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *DeleteTeamContext) AllowTeam() bool {
	return c.allowTeam
}

func (c *DeleteTeamContext) SetIsCaptain(isCaptain bool) {
	c.isCaptain = isCaptain
}

func (c *DeleteTeamContext) IsCaptain() bool {
	return c.isCaptain
}

func (c *DeleteTeamContext) SetMembersCount(count int64) {
	c.membersCount = count
}

func (c *DeleteTeamContext) MembersCount() int64 {
	return c.membersCount
}

type DeleteTeamPolicy struct{}

func NewDeleteTeamPolicy() *DeleteTeamPolicy {
	return &DeleteTeamPolicy{}
}

func (p *DeleteTeamPolicy) Action() policy.Action {
	return ActionDeleteTeam
}

func (p *DeleteTeamPolicy) LoadContext(ctx context.Context, params DeleteTeamParams) (*DeleteTeamContext, error) {
	return NewDeleteTeamContext(), nil
}

func (p *DeleteTeamPolicy) Check(ctx context.Context, pctx *DeleteTeamContext) *policy.Decision {
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
			Message: "teams can only be deleted during REGISTRATION stage",
		})
		return decision
	}

	if !pctx.AllowTeam() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "team deletion is not allowed for this hackathon",
		})
		return decision
	}

	if !pctx.IsCaptain() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only team captain can delete the team",
		})
		return decision
	}

	if pctx.MembersCount() != 1 {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "team can only be deleted when captain is the sole member",
		})
		return decision
	}

	decision.Allow()
	return decision
}
