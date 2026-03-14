package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type SubmitEvaluationParams struct {
	HackathonID  uuid.UUID
	SubmissionID uuid.UUID
}

type SubmitEvaluationContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
	isAssigned     bool
}

func NewSubmitEvaluationContext() *SubmitEvaluationContext {
	return &SubmitEvaluationContext{}
}

func (c *SubmitEvaluationContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *SubmitEvaluationContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *SubmitEvaluationContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *SubmitEvaluationContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *SubmitEvaluationContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *SubmitEvaluationContext) SetIsAssigned(assigned bool) {
	c.isAssigned = assigned
}

func (c *SubmitEvaluationContext) IsJudge() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleJudge {
			return true
		}
	}
	return false
}

type SubmitEvaluationPolicy struct{}

func NewSubmitEvaluationPolicy() *SubmitEvaluationPolicy {
	return &SubmitEvaluationPolicy{}
}

func (p *SubmitEvaluationPolicy) Action() policy.Action {
	return ActionSubmitEvaluation
}

func (p *SubmitEvaluationPolicy) LoadContext(ctx context.Context, params SubmitEvaluationParams) (*SubmitEvaluationContext, error) {
	return NewSubmitEvaluationContext(), nil
}

func (p *SubmitEvaluationPolicy) Check(ctx context.Context, pctx *SubmitEvaluationContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if pctx.hackathonStage != domain.HackathonStageJudging {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeStageRule,
			Message: "evaluations can only be submitted during JUDGING stage",
		})
		return decision
	}

	if !pctx.IsJudge() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only judges can submit evaluations",
		})
		return decision
	}

	if !pctx.isAssigned {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "submission is not assigned to this judge",
		})
		return decision
	}

	decision.Allow()
	return decision
}
