package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type CancelJoinRequestParams struct {
	RequestID uuid.UUID
}

type CancelJoinRequestContext struct {
	authenticated    bool
	actorUserID      uuid.UUID
	hackathonStage   string
	allowTeam        bool
	requestStatus    string
	requestRequester uuid.UUID
}

func NewCancelJoinRequestContext() *CancelJoinRequestContext {
	return &CancelJoinRequestContext{}
}

func (c *CancelJoinRequestContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *CancelJoinRequestContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *CancelJoinRequestContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *CancelJoinRequestContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *CancelJoinRequestContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *CancelJoinRequestContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *CancelJoinRequestContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *CancelJoinRequestContext) AllowTeam() bool {
	return c.allowTeam
}

func (c *CancelJoinRequestContext) SetRequestStatus(status string) {
	c.requestStatus = status
}

func (c *CancelJoinRequestContext) RequestStatus() string {
	return c.requestStatus
}

func (c *CancelJoinRequestContext) SetRequestRequester(requester uuid.UUID) {
	c.requestRequester = requester
}

func (c *CancelJoinRequestContext) RequestRequester() uuid.UUID {
	return c.requestRequester
}

type CancelJoinRequestPolicy struct{}

func NewCancelJoinRequestPolicy() *CancelJoinRequestPolicy {
	return &CancelJoinRequestPolicy{}
}

func (p *CancelJoinRequestPolicy) Action() policy.Action {
	return ActionCancelJoinRequest
}

func (p *CancelJoinRequestPolicy) LoadContext(ctx context.Context, params CancelJoinRequestParams) (*CancelJoinRequestContext, error) {
	return NewCancelJoinRequestContext(), nil
}

func (p *CancelJoinRequestPolicy) Check(ctx context.Context, pctx *CancelJoinRequestContext) *policy.Decision {
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
			Message: "join requests can only be canceled during REGISTRATION stage",
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

	if pctx.RequestStatus() != "pending" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "only pending join requests can be canceled",
		})
		return decision
	}

	if pctx.RequestRequester() != pctx.ActorUserID() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "you can only cancel your own join requests",
		})
		return decision
	}

	decision.Allow()
	return decision
}
