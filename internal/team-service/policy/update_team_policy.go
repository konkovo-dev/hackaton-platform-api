package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type UpdateTeamParams struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
}

type UpdateTeamContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	allowTeam      bool
	isCaptain      bool
}

func NewUpdateTeamContext() *UpdateTeamContext {
	return &UpdateTeamContext{}
}

func (c *UpdateTeamContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *UpdateTeamContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *UpdateTeamContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *UpdateTeamContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *UpdateTeamContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *UpdateTeamContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *UpdateTeamContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *UpdateTeamContext) AllowTeam() bool {
	return c.allowTeam
}

func (c *UpdateTeamContext) SetIsCaptain(isCaptain bool) {
	c.isCaptain = isCaptain
}

func (c *UpdateTeamContext) IsCaptain() bool {
	return c.isCaptain
}

type UpdateTeamPolicy struct{}

func NewUpdateTeamPolicy() *UpdateTeamPolicy {
	return &UpdateTeamPolicy{}
}

func (p *UpdateTeamPolicy) Action() policy.Action {
	return ActionUpdateTeam
}

func (p *UpdateTeamPolicy) LoadContext(ctx context.Context, params UpdateTeamParams) (*UpdateTeamContext, error) {
	return NewUpdateTeamContext(), nil
}

func (p *UpdateTeamPolicy) Check(ctx context.Context, pctx *UpdateTeamContext) *policy.Decision {
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
			Message: "teams can only be updated during REGISTRATION stage",
		})
		return decision
	}

	if !pctx.AllowTeam() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "team updates are not allowed for this hackathon",
		})
		return decision
	}

	if !pctx.IsCaptain() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only team captain can update the team",
		})
		return decision
	}

	decision.Allow()
	return decision
}
