package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type SelectFinalSubmissionParams struct {
	HackathonID     uuid.UUID
	SubmissionID    uuid.UUID
	OwnerKind       string
	OwnerID         uuid.UUID
	CreatedByUserID uuid.UUID
	CaptainUserID   *uuid.UUID
}

type SelectFinalSubmissionContext struct {
	authenticated   bool
	actorUserID     uuid.UUID
	hackathonStage  string
	ownerKind       string
	ownerID         uuid.UUID
	createdByUserID uuid.UUID
	captainUserID   *uuid.UUID
}

func NewSelectFinalSubmissionContext() *SelectFinalSubmissionContext {
	return &SelectFinalSubmissionContext{}
}

func (c *SelectFinalSubmissionContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *SelectFinalSubmissionContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *SelectFinalSubmissionContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *SelectFinalSubmissionContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *SelectFinalSubmissionContext) SetOwner(kind string, id uuid.UUID) {
	c.ownerKind = kind
	c.ownerID = id
}

func (c *SelectFinalSubmissionContext) SetCreatedByUserID(userID uuid.UUID) {
	c.createdByUserID = userID
}

func (c *SelectFinalSubmissionContext) SetCaptainUserID(userID *uuid.UUID) {
	c.captainUserID = userID
}

func (c *SelectFinalSubmissionContext) CanSelectFinal() bool {
	if c.ownerKind == domain.OwnerKindUser {
		return c.actorUserID == c.ownerID
	}

	if c.ownerKind == domain.OwnerKindTeam {
		if c.captainUserID == nil {
			return false
		}
		return c.actorUserID == *c.captainUserID
	}

	return false
}

type SelectFinalSubmissionPolicy struct{}

func NewSelectFinalSubmissionPolicy() *SelectFinalSubmissionPolicy {
	return &SelectFinalSubmissionPolicy{}
}

func (p *SelectFinalSubmissionPolicy) Action() policy.Action {
	return ActionSelectFinalSubmission
}

func (p *SelectFinalSubmissionPolicy) LoadContext(ctx context.Context, params SelectFinalSubmissionParams) (*SelectFinalSubmissionContext, error) {
	pctx := NewSelectFinalSubmissionContext()
	pctx.SetOwner(params.OwnerKind, params.OwnerID)
	pctx.SetCreatedByUserID(params.CreatedByUserID)
	pctx.SetCaptainUserID(params.CaptainUserID)
	return pctx, nil
}

func (p *SelectFinalSubmissionPolicy) Check(ctx context.Context, pctx *SelectFinalSubmissionContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if pctx.hackathonStage != domain.HackathonStageRunning {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeStageRule,
			Message: "final submission can only be selected during RUNNING stage",
		})
		return decision
	}

	if !pctx.CanSelectFinal() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only the owner (or team captain) can select final submission",
		})
		return decision
	}

	decision.Allow()
	return decision
}
