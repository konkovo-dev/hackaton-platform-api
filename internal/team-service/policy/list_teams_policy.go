package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ListTeamsContext struct {
	authenticated       bool
	actorUserID         uuid.UUID
	actorRoles          []string
	participationStatus string
	hackathonStage      string
}

func NewListTeamsContext() *ListTeamsContext {
	return &ListTeamsContext{}
}

func (c *ListTeamsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *ListTeamsContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *ListTeamsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *ListTeamsContext) SetActorUserID(id uuid.UUID) {
	c.actorUserID = id
}

func (c *ListTeamsContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *ListTeamsContext) IsStaff() bool {
	for _, role := range c.actorRoles {
		if role == "owner" || role == "organizer" || role == "mentor" {
			return true
		}
	}
	return false
}

func (c *ListTeamsContext) SetParticipationStatus(status string) {
	c.participationStatus = status
}

func (c *ListTeamsContext) IsParticipant() bool {
	return c.participationStatus != "" && c.participationStatus != "none"
}

func (c *ListTeamsContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *ListTeamsContext) HackathonStage() string {
	return c.hackathonStage
}

type ListTeamsParams struct {
	HackathonID uuid.UUID
}

type ListTeamsPolicy struct{}

func NewListTeamsPolicy() *ListTeamsPolicy {
	return &ListTeamsPolicy{}
}

func (p *ListTeamsPolicy) LoadContext(ctx context.Context, params ListTeamsParams) (*ListTeamsContext, error) {
	return NewListTeamsContext(), nil
}

func (p *ListTeamsPolicy) Check(ctx context.Context, pctx *ListTeamsContext) *policy.Decision {
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
			Message: "teams can only be listed starting from REGISTRATION stage",
		})
		return decision
	}

	if !pctx.IsStaff() && !pctx.IsParticipant() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only staff or participants can list teams",
		})
		return decision
	}

	decision.Allow()
	return decision
}
