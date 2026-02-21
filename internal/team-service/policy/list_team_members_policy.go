package policy

import (
	"context"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ListTeamMembersParams struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
}

type ListTeamMembersContext struct {
	authenticated       bool
	actorUserID         uuid.UUID
	hackathonStage      string
	participationStatus string
	roles               []string
}

func NewListTeamMembersContext() *ListTeamMembersContext {
	return &ListTeamMembersContext{}
}

func (c *ListTeamMembersContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *ListTeamMembersContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *ListTeamMembersContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *ListTeamMembersContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *ListTeamMembersContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *ListTeamMembersContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *ListTeamMembersContext) SetParticipationStatus(status string) {
	c.participationStatus = status
}

func (c *ListTeamMembersContext) ParticipationStatus() string {
	return c.participationStatus
}

func (c *ListTeamMembersContext) SetRoles(roles []string) {
	c.roles = roles
}

func (c *ListTeamMembersContext) Roles() []string {
	return c.roles
}

func (c *ListTeamMembersContext) IsStaff() bool {
	for _, role := range c.roles {
		role = strings.ToLower(role)
		if role == "owner" || role == "organizer" || role == "mentor" {
			return true
		}
	}
	return false
}

func (c *ListTeamMembersContext) IsParticipant() bool {
	status := strings.ToLower(c.participationStatus)
	return status != "" && status != "part_none"
}

type ListTeamMembersPolicy struct{}

func NewListTeamMembersPolicy() *ListTeamMembersPolicy {
	return &ListTeamMembersPolicy{}
}

func (p *ListTeamMembersPolicy) Action() policy.Action {
	return ActionListTeamMembers
}

func (p *ListTeamMembersPolicy) LoadContext(ctx context.Context, params ListTeamMembersParams) (*ListTeamMembersContext, error) {
	return NewListTeamMembersContext(), nil
}

func (p *ListTeamMembersPolicy) Check(ctx context.Context, pctx *ListTeamMembersContext) *policy.Decision {
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
			Message: "team members can only be viewed during REGISTRATION, PRESTART, RUNNING, JUDGING, or FINISHED stages",
		})
		return decision
	}

	if !pctx.IsStaff() && !pctx.IsParticipant() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "caller must be staff or a registered participant",
		})
		return decision
	}

	decision.Allow()
	return decision
}
