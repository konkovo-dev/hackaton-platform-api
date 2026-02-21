package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type UpsertVacancyParams struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
}

type UpsertVacancyContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	allowTeam      bool
	isCaptain      bool
}

func NewUpsertVacancyContext() *UpsertVacancyContext {
	return &UpsertVacancyContext{}
}

func (c *UpsertVacancyContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *UpsertVacancyContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *UpsertVacancyContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *UpsertVacancyContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *UpsertVacancyContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *UpsertVacancyContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *UpsertVacancyContext) SetAllowTeam(allow bool) {
	c.allowTeam = allow
}

func (c *UpsertVacancyContext) AllowTeam() bool {
	return c.allowTeam
}

func (c *UpsertVacancyContext) SetIsCaptain(isCaptain bool) {
	c.isCaptain = isCaptain
}

func (c *UpsertVacancyContext) IsCaptain() bool {
	return c.isCaptain
}

type UpsertVacancyPolicy struct{}

func NewUpsertVacancyPolicy() *UpsertVacancyPolicy {
	return &UpsertVacancyPolicy{}
}

func (p *UpsertVacancyPolicy) Action() policy.Action {
	return ActionUpsertVacancy
}

func (p *UpsertVacancyPolicy) LoadContext(ctx context.Context, params UpsertVacancyParams) (*UpsertVacancyContext, error) {
	return NewUpsertVacancyContext(), nil
}

func (p *UpsertVacancyPolicy) Check(ctx context.Context, pctx *UpsertVacancyContext) *policy.Decision {
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
			Message: "vacancies can only be managed during REGISTRATION stage",
		})
		return decision
	}

	if !pctx.AllowTeam() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodePolicyRule,
			Message: "vacancy management is not allowed for this hackathon",
		})
		return decision
	}

	if !pctx.IsCaptain() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only team captain can manage vacancies",
		})
		return decision
	}

	decision.Allow()
	return decision
}
