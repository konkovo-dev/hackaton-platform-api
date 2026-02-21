package policy

import (
	"context"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type CreateJoinRequestParams struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
	VacancyID   uuid.UUID
}

type CreateJoinRequestContext struct {
	authenticated       bool
	actorUserID         uuid.UUID
	hackathonStage      string
	allowTeam           bool
	teamIsJoinable      bool
	isStaff             bool
	participationStatus string
	slotsOpen           int64
}

func NewCreateJoinRequestContext() *CreateJoinRequestContext {
	return &CreateJoinRequestContext{}
}

func (c *CreateJoinRequestContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *CreateJoinRequestContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *CreateJoinRequestContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *CreateJoinRequestContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *CreateJoinRequestContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *CreateJoinRequestContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *CreateJoinRequestContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *CreateJoinRequestContext) AllowTeam() bool {
	return c.allowTeam
}

func (c *CreateJoinRequestContext) SetTeamIsJoinable(isJoinable bool) {
	c.teamIsJoinable = isJoinable
}

func (c *CreateJoinRequestContext) TeamIsJoinable() bool {
	return c.teamIsJoinable
}

func (c *CreateJoinRequestContext) SetIsStaff(isStaff bool) {
	c.isStaff = isStaff
}

func (c *CreateJoinRequestContext) IsStaff() bool {
	return c.isStaff
}

func (c *CreateJoinRequestContext) SetParticipationStatus(status string) {
	c.participationStatus = status
}

func (c *CreateJoinRequestContext) ParticipationStatus() string {
	return c.participationStatus
}

func (c *CreateJoinRequestContext) CanSeekTeam() bool {
	status := strings.ToLower(c.participationStatus)
	return status == "looking_for_team" || status == "individual"
}

func (c *CreateJoinRequestContext) IsTeamMember() bool {
	status := strings.ToLower(c.participationStatus)
	return status == "team_member" || status == "team_captain"
}

func (c *CreateJoinRequestContext) SetSlotsOpen(slots int64) {
	c.slotsOpen = slots
}

func (c *CreateJoinRequestContext) SlotsOpen() int64 {
	return c.slotsOpen
}

type CreateJoinRequestPolicy struct{}

func NewCreateJoinRequestPolicy() *CreateJoinRequestPolicy {
	return &CreateJoinRequestPolicy{}
}

func (p *CreateJoinRequestPolicy) Action() policy.Action {
	return ActionCreateJoinRequest
}

func (p *CreateJoinRequestPolicy) LoadContext(ctx context.Context, params CreateJoinRequestParams) (*CreateJoinRequestContext, error) {
	return NewCreateJoinRequestContext(), nil
}

func (p *CreateJoinRequestPolicy) Check(ctx context.Context, pctx *CreateJoinRequestContext) *policy.Decision {
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
			Message: "join requests can only be created during REGISTRATION stage",
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

	if !pctx.TeamIsJoinable() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "team is not accepting join requests",
		})
		return decision
	}

	if !pctx.CanSeekTeam() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "only users looking for team or participating individually can send join requests",
		})
		return decision
	}

	if pctx.IsStaff() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "staff members cannot join teams",
		})
		return decision
	}

	if pctx.IsTeamMember() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "you are already in a team",
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

	decision.Allow()
	return decision
}
