package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type TransferCaptainParams struct {
	HackathonID    uuid.UUID
	TeamID         uuid.UUID
	NewCaptainID   uuid.UUID
}

type TransferCaptainContext struct {
	authenticated      bool
	actorUserID        uuid.UUID
	hackathonStage     string
	allowTeam          bool
	isCaptain          bool
	targetIsMember     bool
	newCaptainID       uuid.UUID
}

func NewTransferCaptainContext() *TransferCaptainContext {
	return &TransferCaptainContext{}
}

func (c *TransferCaptainContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *TransferCaptainContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *TransferCaptainContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *TransferCaptainContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *TransferCaptainContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *TransferCaptainContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *TransferCaptainContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *TransferCaptainContext) AllowTeam() bool {
	return c.allowTeam
}

func (c *TransferCaptainContext) SetIsCaptain(isCaptain bool) {
	c.isCaptain = isCaptain
}

func (c *TransferCaptainContext) IsCaptain() bool {
	return c.isCaptain
}

func (c *TransferCaptainContext) SetTargetIsMember(isMember bool) {
	c.targetIsMember = isMember
}

func (c *TransferCaptainContext) TargetIsMember() bool {
	return c.targetIsMember
}

func (c *TransferCaptainContext) SetNewCaptainID(userID uuid.UUID) {
	c.newCaptainID = userID
}

func (c *TransferCaptainContext) NewCaptainID() uuid.UUID {
	return c.newCaptainID
}

type TransferCaptainPolicy struct{}

func NewTransferCaptainPolicy() *TransferCaptainPolicy {
	return &TransferCaptainPolicy{}
}

func (p *TransferCaptainPolicy) Action() policy.Action {
	return ActionTransferCaptain
}

func (p *TransferCaptainPolicy) LoadContext(ctx context.Context, params TransferCaptainParams) (*TransferCaptainContext, error) {
	return NewTransferCaptainContext(), nil
}

func (p *TransferCaptainPolicy) Check(ctx context.Context, pctx *TransferCaptainContext) *policy.Decision {
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
			Message: "only team captain can transfer captaincy",
		})
		return decision
	}

	if pctx.HackathonStage() != "registration" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeStageRule,
			Message: "captaincy can only be transferred during REGISTRATION stage",
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

	if !pctx.TargetIsMember() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeNotFound,
			Message: "target user is not a team member",
		})
		return decision
	}

	if pctx.NewCaptainID() == pctx.ActorUserID() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "cannot transfer captaincy to yourself",
		})
		return decision
	}

	decision.Allow()
	return decision
}
