package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type RegisterForHackathonParams struct {
	HackathonID   uuid.UUID
	DesiredStatus string
}

type RegisterParticipationContext struct {
	authenticated bool
	actorUserID   uuid.UUID

	actorRoles              []string
	actorParticipationKind  string
	desiredStatus           string
	hackathonStage          string
	allowIndividual         bool
	allowTeam               bool
}

func NewRegisterParticipationContext() *RegisterParticipationContext {
	return &RegisterParticipationContext{}
}

func (c *RegisterParticipationContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *RegisterParticipationContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *RegisterParticipationContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *RegisterParticipationContext) SetActorUserID(id uuid.UUID) {
	c.actorUserID = id
}

func (c *RegisterParticipationContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *RegisterParticipationContext) ActorHasRole(role string) bool {
	for _, r := range c.actorRoles {
		if r == role {
			return true
		}
	}
	return false
}

func (c *RegisterParticipationContext) ActorHasAnyStaffRole() bool {
	return len(c.actorRoles) > 0
}

func (c *RegisterParticipationContext) SetActorParticipationKind(kind string) {
	c.actorParticipationKind = kind
}

func (c *RegisterParticipationContext) ActorParticipationKind() string {
	return c.actorParticipationKind
}

func (c *RegisterParticipationContext) SetDesiredStatus(status string) {
	c.desiredStatus = status
}

func (c *RegisterParticipationContext) DesiredStatus() string {
	return c.desiredStatus
}

func (c *RegisterParticipationContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *RegisterParticipationContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *RegisterParticipationContext) SetAllowIndividual(allow bool) {
	c.allowIndividual = allow
}

func (c *RegisterParticipationContext) AllowIndividual() bool {
	return c.allowIndividual
}

func (c *RegisterParticipationContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *RegisterParticipationContext) AllowTeam() bool {
	return c.allowTeam
}

type HackathonInfo struct {
	Stage           string
	AllowIndividual bool
	AllowTeam       bool
}

type RegisterForHackathonRepositories interface {
	GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error)
	GetParticipationStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error)
	GetHackathonInfo(ctx context.Context, hackathonID uuid.UUID) (*HackathonInfo, error)
}

type RegisterForHackathonPolicy struct {
	policy.BasePolicy
	repos RegisterForHackathonRepositories
}

func NewRegisterForHackathonPolicy(repos RegisterForHackathonRepositories) *RegisterForHackathonPolicy {
	return &RegisterForHackathonPolicy{
		BasePolicy: policy.NewBasePolicy(ActionRegisterForHackathon),
		repos:      repos,
	}
}

func (p *RegisterForHackathonPolicy) LoadContext(ctx context.Context, params RegisterForHackathonParams) (policy.PolicyContext, error) {
	pctx := NewRegisterParticipationContext()
	pctx.SetDesiredStatus(params.DesiredStatus)

	// Load hackathon info (stage, allow_individual, allow_team)
	hackathonInfo, err := p.repos.GetHackathonInfo(ctx, params.HackathonID)
	if err != nil {
		return nil, err
	}
	pctx.SetHackathonStage(hackathonInfo.Stage)
	pctx.SetAllowIndividual(hackathonInfo.AllowIndividual)
	pctx.SetAllowTeam(hackathonInfo.AllowTeam)

	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return pctx, nil
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return pctx, nil
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)

	actorRoles, err := p.repos.GetRoleStrings(ctx, params.HackathonID, userUUID)
	if err == nil {
		pctx.SetActorRoles(actorRoles)
	}

	actorParticipationStatus, err := p.repos.GetParticipationStatus(ctx, params.HackathonID, userUUID)
	if err == nil {
		pctx.SetActorParticipationKind(actorParticipationStatus)
	}

	return pctx, nil
}

func (p *RegisterForHackathonPolicy) Check(ctx context.Context, pctx policy.PolicyContext) *policy.Decision {
	decision := policy.NewDecision()
	regCtx := pctx.(*RegisterParticipationContext)

	if !regCtx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "authentication required",
		})
		return decision
	}

	// Check ParticipationWriteWindow: stage must be REGISTRATION
	if regCtx.HackathonStage() != "registration" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "registration is only allowed during REGISTRATION stage",
		})
		return decision
	}

	if regCtx.ActorHasAnyStaffRole() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "staff members cannot register as participants",
		})
		return decision
	}

	participationKind := regCtx.ActorParticipationKind()
	if participationKind != "" && participationKind != string(domain.ParticipationNone) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeConflict,
			Message: "user is already registered for this hackathon",
		})
		return decision
	}

	desiredStatus := regCtx.DesiredStatus()
	if desiredStatus != string(domain.ParticipationIndividual) &&
		desiredStatus != string(domain.ParticipationLookingForTeam) {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "can only register with INDIVIDUAL_ACTIVE or LOOKING_FOR_TEAM status",
		})
		return decision
	}

	// Check registration policy: if registering as INDIVIDUAL, allow_individual must be true
	if desiredStatus == string(domain.ParticipationIndividual) && !regCtx.AllowIndividual() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "individual registration is not allowed for this hackathon",
		})
		return decision
	}

	// Check registration policy: if registering as LOOKING_FOR_TEAM, allow_team must be true
	if desiredStatus == string(domain.ParticipationLookingForTeam) && !regCtx.AllowTeam() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "team registration is not allowed for this hackathon",
		})
		return decision
	}

	return decision
}

var _ policy.Policy[RegisterForHackathonParams] = (*RegisterForHackathonPolicy)(nil)
