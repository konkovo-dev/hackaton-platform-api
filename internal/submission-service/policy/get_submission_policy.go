package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type GetSubmissionParams struct {
	HackathonID uuid.UUID
	SubmissionID uuid.UUID
	OwnerKind   string
	OwnerID     uuid.UUID
	ActorOwnerKind string
	ActorOwnerID   uuid.UUID
}

type GetSubmissionContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	actorRoles     []string
	hackathonStage string
	ownerKind      string
	ownerID        uuid.UUID
	actorOwnerKind string
	actorOwnerID   uuid.UUID
}

func NewGetSubmissionContext() *GetSubmissionContext {
	return &GetSubmissionContext{}
}

func (c *GetSubmissionContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *GetSubmissionContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *GetSubmissionContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *GetSubmissionContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *GetSubmissionContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *GetSubmissionContext) SetOwner(kind string, id uuid.UUID) {
	c.ownerKind = kind
	c.ownerID = id
}

func (c *GetSubmissionContext) SetActorOwner(kind string, id uuid.UUID) {
	c.actorOwnerKind = kind
	c.actorOwnerID = id
}

func (c *GetSubmissionContext) IsStaff() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleOwner || role == domain.RoleOrganizer || role == domain.RoleMentor || role == domain.RoleJudge {
			return true
		}
	}
	return false
}

func (c *GetSubmissionContext) IsOwner() bool {
	return c.ownerKind == c.actorOwnerKind && c.ownerID == c.actorOwnerID
}

func (c *GetSubmissionContext) IsReadWindowStage() bool {
	return c.hackathonStage == domain.HackathonStageRunning ||
		c.hackathonStage == domain.HackathonStageJudging ||
		c.hackathonStage == domain.HackathonStageFinished
}

type GetSubmissionPolicy struct{}

func NewGetSubmissionPolicy() *GetSubmissionPolicy {
	return &GetSubmissionPolicy{}
}

func (p *GetSubmissionPolicy) Action() policy.Action {
	return ActionGetSubmission
}

func (p *GetSubmissionPolicy) LoadContext(ctx context.Context, params GetSubmissionParams) (*GetSubmissionContext, error) {
	pctx := NewGetSubmissionContext()
	pctx.SetOwner(params.OwnerKind, params.OwnerID)
	pctx.SetActorOwner(params.ActorOwnerKind, params.ActorOwnerID)
	return pctx, nil
}

func (p *GetSubmissionPolicy) Check(ctx context.Context, pctx *GetSubmissionContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if !pctx.IsReadWindowStage() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeStageRule,
			Message: "submissions can only be viewed during RUNNING, JUDGING, or FINISHED stages",
		})
		return decision
	}

	if !pctx.IsOwner() && !pctx.IsStaff() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only owner or staff can view this submission",
		})
		return decision
	}

	decision.Allow()
	return decision
}
