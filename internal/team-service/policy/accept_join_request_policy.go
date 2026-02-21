package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type AcceptJoinRequestParams struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
	RequestID   uuid.UUID
}

type AcceptJoinRequestContext struct {
	authenticated       bool
	actorUserID         uuid.UUID
	hackathonStage      string
	allowTeam           bool
	isCaptain           bool
	requestStatus       string
	requesterIsStaff    bool
	requesterIsTeamMember bool
	slotsOpen           int64
}

func NewAcceptJoinRequestContext() *AcceptJoinRequestContext {
	return &AcceptJoinRequestContext{}
}

func (c *AcceptJoinRequestContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *AcceptJoinRequestContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *AcceptJoinRequestContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *AcceptJoinRequestContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *AcceptJoinRequestContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *AcceptJoinRequestContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *AcceptJoinRequestContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *AcceptJoinRequestContext) AllowTeam() bool {
	return c.allowTeam
}

func (c *AcceptJoinRequestContext) SetIsCaptain(isCaptain bool) {
	c.isCaptain = isCaptain
}

func (c *AcceptJoinRequestContext) IsCaptain() bool {
	return c.isCaptain
}

func (c *AcceptJoinRequestContext) SetRequestStatus(status string) {
	c.requestStatus = status
}

func (c *AcceptJoinRequestContext) RequestStatus() string {
	return c.requestStatus
}

func (c *AcceptJoinRequestContext) SetRequesterIsStaff(isStaff bool) {
	c.requesterIsStaff = isStaff
}

func (c *AcceptJoinRequestContext) RequesterIsStaff() bool {
	return c.requesterIsStaff
}

func (c *AcceptJoinRequestContext) SetRequesterIsTeamMember(isTeamMember bool) {
	c.requesterIsTeamMember = isTeamMember
}

func (c *AcceptJoinRequestContext) RequesterIsTeamMember() bool {
	return c.requesterIsTeamMember
}

func (c *AcceptJoinRequestContext) SetSlotsOpen(slots int64) {
	c.slotsOpen = slots
}

func (c *AcceptJoinRequestContext) SlotsOpen() int64 {
	return c.slotsOpen
}

type AcceptJoinRequestPolicy struct{}

func NewAcceptJoinRequestPolicy() *AcceptJoinRequestPolicy {
	return &AcceptJoinRequestPolicy{}
}

func (p *AcceptJoinRequestPolicy) Action() policy.Action {
	return ActionAcceptJoinRequest
}

func (p *AcceptJoinRequestPolicy) LoadContext(ctx context.Context, params AcceptJoinRequestParams) (*AcceptJoinRequestContext, error) {
	return NewAcceptJoinRequestContext(), nil
}

func (p *AcceptJoinRequestPolicy) Check(ctx context.Context, pctx *AcceptJoinRequestContext) *policy.Decision {
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
			Message: "only team captain can accept join requests",
		})
		return decision
	}

	if pctx.HackathonStage() != "registration" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeStageRule,
			Message: "join requests can only be accepted during REGISTRATION stage",
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
			Message: "only pending join requests can be accepted",
		})
		return decision
	}

	if pctx.SlotsOpen() <= 0 {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "no open slots in vacancy",
		})
		return decision
	}

	if pctx.RequesterIsStaff() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "staff members cannot join teams",
		})
		return decision
	}

	if pctx.RequesterIsTeamMember() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "requester is already in a team",
		})
		return decision
	}

	decision.Allow()
	return decision
}
