package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type GetMyEvaluationsParams struct {
	HackathonID uuid.UUID
}

type GetMyEvaluationsContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
}

func NewGetMyEvaluationsContext() *GetMyEvaluationsContext {
	return &GetMyEvaluationsContext{}
}

func (c *GetMyEvaluationsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *GetMyEvaluationsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *GetMyEvaluationsContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *GetMyEvaluationsContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *GetMyEvaluationsContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *GetMyEvaluationsContext) IsJudge() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleJudge {
			return true
		}
	}
	return false
}

func (c *GetMyEvaluationsContext) IsJudgingOrLater() bool {
	return c.hackathonStage == domain.HackathonStageJudging ||
		c.hackathonStage == domain.HackathonStageFinished
}

type GetMyEvaluationsPolicy struct{}

func NewGetMyEvaluationsPolicy() *GetMyEvaluationsPolicy {
	return &GetMyEvaluationsPolicy{}
}

func (p *GetMyEvaluationsPolicy) Action() policy.Action {
	return ActionGetMyEvaluations
}

func (p *GetMyEvaluationsPolicy) LoadContext(ctx context.Context, params GetMyEvaluationsParams) (*GetMyEvaluationsContext, error) {
	return NewGetMyEvaluationsContext(), nil
}

func (p *GetMyEvaluationsPolicy) Check(ctx context.Context, pctx *GetMyEvaluationsContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if !pctx.IsJudgingOrLater() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeStageRule,
			Message: "evaluations can only be viewed from JUDGING stage onwards",
		})
		return decision
	}

	if !pctx.IsJudge() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only judges can view their evaluations",
		})
		return decision
	}

	decision.Allow()
	return decision
}
