package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type GetMyEvaluationResultParams struct {
	HackathonID uuid.UUID
}

type GetMyEvaluationResultContext struct {
	authenticated      bool
	actorUserID        uuid.UUID
	resultPublishedAt  bool
	participationStatus string
}

func NewGetMyEvaluationResultContext() *GetMyEvaluationResultContext {
	return &GetMyEvaluationResultContext{}
}

func (c *GetMyEvaluationResultContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *GetMyEvaluationResultContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *GetMyEvaluationResultContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *GetMyEvaluationResultContext) SetResultPublishedAt(published bool) {
	c.resultPublishedAt = published
}

func (c *GetMyEvaluationResultContext) SetParticipationStatus(status string) {
	c.participationStatus = status
}

type GetMyEvaluationResultPolicy struct{}

func NewGetMyEvaluationResultPolicy() *GetMyEvaluationResultPolicy {
	return &GetMyEvaluationResultPolicy{}
}

func (p *GetMyEvaluationResultPolicy) Action() policy.Action {
	return ActionGetMyEvaluationResult
}

func (p *GetMyEvaluationResultPolicy) LoadContext(ctx context.Context, params GetMyEvaluationResultParams) (*GetMyEvaluationResultContext, error) {
	return NewGetMyEvaluationResultContext(), nil
}

func (p *GetMyEvaluationResultPolicy) Check(ctx context.Context, pctx *GetMyEvaluationResultContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if !pctx.resultPublishedAt {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "results have not been published yet",
		})
		return decision
	}

	decision.Allow()
	return decision
}
