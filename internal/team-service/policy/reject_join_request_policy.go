package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type RejectJoinRequestParams struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
	RequestID   uuid.UUID
}

type RejectJoinRequestContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	allowTeam      bool
	isCaptain      bool
	requestStatus  string
}

func NewRejectJoinRequestContext() *RejectJoinRequestContext {
	return &RejectJoinRequestContext{}
}

func (c *RejectJoinRequestContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *RejectJoinRequestContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *RejectJoinRequestContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *RejectJoinRequestContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *RejectJoinRequestContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *RejectJoinRequestContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *RejectJoinRequestContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *RejectJoinRequestContext) AllowTeam() bool {
	return c.allowTeam
}

func (c *RejectJoinRequestContext) SetIsCaptain(isCaptain bool) {
	c.isCaptain = isCaptain
}

func (c *RejectJoinRequestContext) IsCaptain() bool {
	return c.isCaptain
}

func (c *RejectJoinRequestContext) SetRequestStatus(status string) {
	c.requestStatus = status
}

func (c *RejectJoinRequestContext) RequestStatus() string {
	return c.requestStatus
}

type RejectJoinRequestPolicy struct{}

func NewRejectJoinRequestPolicy() *RejectJoinRequestPolicy {
	return &RejectJoinRequestPolicy{}
}

func (p *RejectJoinRequestPolicy) Action() policy.Action {
	return ActionRejectJoinRequest
}

func (p *RejectJoinRequestPolicy) LoadContext(ctx context.Context, params RejectJoinRequestParams) (*RejectJoinRequestContext, error) {
	return NewRejectJoinRequestContext(), nil
}

func (p *RejectJoinRequestPolicy) Check(ctx context.Context, pctx *RejectJoinRequestContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if !pctx.IsCaptain() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only team captain can reject join requests",
		})
		return decision
	}

	if pctx.HackathonStage() != "registration" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeStageRule,
			Message: "join requests can only be rejected during REGISTRATION stage",
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
			Message: "only pending join requests can be rejected",
		})
		return decision
	}

	decision.Allow()
	return decision
}
