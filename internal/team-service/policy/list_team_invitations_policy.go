package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ListTeamInvitationsParams struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
}

type ListTeamInvitationsContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	isCaptain      bool
}

func NewListTeamInvitationsContext() *ListTeamInvitationsContext {
	return &ListTeamInvitationsContext{}
}

func (c *ListTeamInvitationsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *ListTeamInvitationsContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *ListTeamInvitationsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *ListTeamInvitationsContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *ListTeamInvitationsContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *ListTeamInvitationsContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *ListTeamInvitationsContext) SetIsCaptain(isCaptain bool) {
	c.isCaptain = isCaptain
}

func (c *ListTeamInvitationsContext) IsCaptain() bool {
	return c.isCaptain
}

type ListTeamInvitationsPolicy struct{}

func NewListTeamInvitationsPolicy() *ListTeamInvitationsPolicy {
	return &ListTeamInvitationsPolicy{}
}

func (p *ListTeamInvitationsPolicy) Action() policy.Action {
	return ActionListTeamInvitations
}

func (p *ListTeamInvitationsPolicy) LoadContext(ctx context.Context, params ListTeamInvitationsParams) (*ListTeamInvitationsContext, error) {
	return NewListTeamInvitationsContext(), nil
}

func (p *ListTeamInvitationsPolicy) Check(ctx context.Context, pctx *ListTeamInvitationsContext) *policy.Decision {
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
			Message: "team invitations can only be viewed during REGISTRATION, PRESTART, RUNNING, JUDGING, or FINISHED stages",
		})
		return decision
	}

	if !pctx.IsCaptain() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only team captain can view team invitations",
		})
		return decision
	}

	decision.Allow()
	return decision
}
