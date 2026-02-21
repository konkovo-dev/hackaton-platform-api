package policy

import (
	"context"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ListJoinRequestsParams struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
}

type ListJoinRequestsContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	isCaptain      bool
}

func NewListJoinRequestsContext() *ListJoinRequestsContext {
	return &ListJoinRequestsContext{}
}

func (c *ListJoinRequestsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *ListJoinRequestsContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *ListJoinRequestsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *ListJoinRequestsContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *ListJoinRequestsContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *ListJoinRequestsContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *ListJoinRequestsContext) SetIsCaptain(isCaptain bool) {
	c.isCaptain = isCaptain
}

func (c *ListJoinRequestsContext) IsCaptain() bool {
	return c.isCaptain
}

func (c *ListJoinRequestsContext) IsInTeamReadWindow() bool {
	stage := strings.ToLower(c.hackathonStage)
	return stage == "registration" || stage == "prestart" || stage == "running" || stage == "judging" || stage == "finished"
}

type ListJoinRequestsPolicy struct{}

func NewListJoinRequestsPolicy() *ListJoinRequestsPolicy {
	return &ListJoinRequestsPolicy{}
}

func (p *ListJoinRequestsPolicy) Action() policy.Action {
	return ActionListJoinRequests
}

func (p *ListJoinRequestsPolicy) LoadContext(ctx context.Context, params ListJoinRequestsParams) (*ListJoinRequestsContext, error) {
	return NewListJoinRequestsContext(), nil
}

func (p *ListJoinRequestsPolicy) Check(ctx context.Context, pctx *ListJoinRequestsContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if !pctx.IsInTeamReadWindow() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeStageRule,
			Message: "join requests can only be viewed during REGISTRATION, PRESTART, RUNNING, JUDGING, or FINISHED stages",
		})
		return decision
	}

	if !pctx.IsCaptain() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only team captain can view join requests",
		})
		return decision
	}

	decision.Allow()
	return decision
}
