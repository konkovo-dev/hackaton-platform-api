package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type GetSubmissionEvaluationsParams struct {
	HackathonID  uuid.UUID
	SubmissionID uuid.UUID
}

type GetSubmissionEvaluationsContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
}

func NewGetSubmissionEvaluationsContext() *GetSubmissionEvaluationsContext {
	return &GetSubmissionEvaluationsContext{}
}

func (c *GetSubmissionEvaluationsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *GetSubmissionEvaluationsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *GetSubmissionEvaluationsContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *GetSubmissionEvaluationsContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *GetSubmissionEvaluationsContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *GetSubmissionEvaluationsContext) IsStaffOrJudge() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleOwner || role == domain.RoleOrganizer || role == domain.RoleJudge {
			return true
		}
	}
	return false
}

func (c *GetSubmissionEvaluationsContext) IsJudgingOrLater() bool {
	return c.hackathonStage == domain.HackathonStageJudging ||
		c.hackathonStage == domain.HackathonStageFinished
}

type GetSubmissionEvaluationsPolicy struct{}

func NewGetSubmissionEvaluationsPolicy() *GetSubmissionEvaluationsPolicy {
	return &GetSubmissionEvaluationsPolicy{}
}

func (p *GetSubmissionEvaluationsPolicy) Action() policy.Action {
	return ActionGetSubmissionEvaluations
}

func (p *GetSubmissionEvaluationsPolicy) LoadContext(ctx context.Context, params GetSubmissionEvaluationsParams) (*GetSubmissionEvaluationsContext, error) {
	return NewGetSubmissionEvaluationsContext(), nil
}

func (p *GetSubmissionEvaluationsPolicy) Check(ctx context.Context, pctx *GetSubmissionEvaluationsContext) *policy.Decision {
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
			Message: "submission evaluations can only be viewed from JUDGING stage onwards",
		})
		return decision
	}

	if !pctx.IsStaffOrJudge() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only organizers and judges can view submission evaluations",
		})
		return decision
	}

	decision.Allow()
	return decision
}
