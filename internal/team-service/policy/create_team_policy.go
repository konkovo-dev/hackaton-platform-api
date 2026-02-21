package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type CreateTeamParams struct {
	HackathonID uuid.UUID
}

type CreateTeamContext struct {
	authenticated       bool
	actorUserID         uuid.UUID
	actorRoles          []string
	participationStatus string
	hackathonStage      string
	allowTeam           bool
}

func NewCreateTeamContext() *CreateTeamContext {
	return &CreateTeamContext{}
}

func (c *CreateTeamContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *CreateTeamContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *CreateTeamContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *CreateTeamContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *CreateTeamContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *CreateTeamContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *CreateTeamContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *CreateTeamContext) ActorRoles() []string {
	return c.actorRoles
}

func (c *CreateTeamContext) SetParticipationStatus(status string) {
	c.participationStatus = status
}

func (c *CreateTeamContext) ParticipationStatus() string {
	return c.participationStatus
}

func (c *CreateTeamContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *CreateTeamContext) AllowTeam() bool {
	return c.allowTeam
}

func (c *CreateTeamContext) IsStaff() bool {
	for _, role := range c.actorRoles {
		if role == "owner" || role == "organizer" || role == "mentor" {
			return true
		}
	}
	return false
}

func (c *CreateTeamContext) IsTeamMember() bool {
	return c.participationStatus == "team_member" || c.participationStatus == "team_captain"
}

type CreateTeamPolicy struct{}

func NewCreateTeamPolicy() *CreateTeamPolicy {
	return &CreateTeamPolicy{}
}

func (p *CreateTeamPolicy) Action() policy.Action {
	return ActionCreateTeam
}

func (p *CreateTeamPolicy) LoadContext(ctx context.Context, params CreateTeamParams) (*CreateTeamContext, error) {
	return NewCreateTeamContext(), nil
}

func (p *CreateTeamPolicy) Check(ctx context.Context, pctx *CreateTeamContext) *policy.Decision {
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
			Message: "teams can only be created during REGISTRATION stage",
		})
		return decision
	}

	if !pctx.AllowTeam() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "team creation is not allowed for this hackathon",
		})
		return decision
	}

	if pctx.IsStaff() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "staff cannot create teams",
		})
		return decision
	}

	if pctx.IsTeamMember() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "user is already a member of a team",
		})
		return decision
	}

	decision.Allow()
	return decision
}
