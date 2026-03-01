package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type UpdateSubmissionParams struct {
	HackathonID     uuid.UUID
	SubmissionID    uuid.UUID
	OwnerKind       string
	OwnerID         uuid.UUID
	CreatedByUserID uuid.UUID
}

type UpdateSubmissionContext struct {
	authenticated   bool
	actorUserID     uuid.UUID
	actorOwnerKind  string
	actorOwnerID    uuid.UUID
	hackathonStage  string
	ownerKind       string
	ownerID         uuid.UUID
	createdByUserID uuid.UUID
}

func NewUpdateSubmissionContext() *UpdateSubmissionContext {
	return &UpdateSubmissionContext{}
}

func (c *UpdateSubmissionContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *UpdateSubmissionContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *UpdateSubmissionContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *UpdateSubmissionContext) SetActorOwnerKind(kind string) {
	c.actorOwnerKind = kind
}

func (c *UpdateSubmissionContext) SetActorOwnerID(id uuid.UUID) {
	c.actorOwnerID = id
}

func (c *UpdateSubmissionContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *UpdateSubmissionContext) SetOwnerKind(kind string) {
	c.ownerKind = kind
}

func (c *UpdateSubmissionContext) SetOwnerID(id uuid.UUID) {
	c.ownerID = id
}

func (c *UpdateSubmissionContext) SetCreatedByUserID(userID uuid.UUID) {
	c.createdByUserID = userID
}

func (c *UpdateSubmissionContext) IsCreator() bool {
	return c.actorUserID == c.createdByUserID
}

func (c *UpdateSubmissionContext) IsOwner() bool {
	return c.actorOwnerKind == c.ownerKind && c.actorOwnerID == c.ownerID
}

type UpdateSubmissionPolicy struct{}

func NewUpdateSubmissionPolicy() *UpdateSubmissionPolicy {
	return &UpdateSubmissionPolicy{}
}

func (p *UpdateSubmissionPolicy) Action() policy.Action {
	return ActionUpdateSubmission
}

func (p *UpdateSubmissionPolicy) LoadContext(ctx context.Context, params UpdateSubmissionParams) (*UpdateSubmissionContext, error) {
	pctx := NewUpdateSubmissionContext()
	pctx.SetOwnerKind(params.OwnerKind)
	pctx.SetOwnerID(params.OwnerID)
	pctx.SetCreatedByUserID(params.CreatedByUserID)
	return pctx, nil
}

func (p *UpdateSubmissionPolicy) Check(ctx context.Context, pctx *UpdateSubmissionContext) *policy.Decision {
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
			Message: "submissions can only be updated during RUNNING stage",
		})
		return decision
	}

	if !pctx.IsOwner() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only the owner can update the submission",
		})
		return decision
	}

	decision.Allow()
	return decision
}
