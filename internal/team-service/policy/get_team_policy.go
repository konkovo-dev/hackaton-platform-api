package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type GetTeamParams struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
}

type GetTeamContext struct {
	authenticated       bool
	actorUserID         uuid.UUID
	actorRoles          []string
	participationStatus string
	hackathonStage      string
}

func NewGetTeamContext() *GetTeamContext {
	return &GetTeamContext{}
}

func (c *GetTeamContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *GetTeamContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *GetTeamContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *GetTeamContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *GetTeamContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *GetTeamContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *GetTeamContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *GetTeamContext) ActorRoles() []string {
	return c.actorRoles
}

func (c *GetTeamContext) SetParticipationStatus(status string) {
	c.participationStatus = status
}

func (c *GetTeamContext) ParticipationStatus() string {
	return c.participationStatus
}

func (c *GetTeamContext) IsStaff() bool {
	for _, role := range c.actorRoles {
		if role == "owner" || role == "organizer" || role == "mentor" {
			return true
		}
	}
	return false
}

func (c *GetTeamContext) IsParticipant() bool {
	return c.participationStatus != "" && c.participationStatus != "none"
}

type GetTeamPolicy struct{}

func NewGetTeamPolicy() *GetTeamPolicy {
	return &GetTeamPolicy{}
}

func (p *GetTeamPolicy) Action() policy.Action {
	return ActionGetTeam
}

func (p *GetTeamPolicy) LoadContext(ctx context.Context, params GetTeamParams) (*GetTeamContext, error) {
	return NewGetTeamContext(), nil
}

func (p *GetTeamPolicy) Check(ctx context.Context, pctx *GetTeamContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	stage := pctx.HackathonStage()
	if stage != "registration" && stage != "prestart" && stage != "running" && stage != "judging" && stage != "finished" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeStageRule,
			Message: "teams can only be viewed starting from REGISTRATION stage",
		})
		return decision
	}

	if !pctx.IsStaff() && !pctx.IsParticipant() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only staff or participants can view teams",
		})
		return decision
	}

	decision.Allow()
	return decision
}
