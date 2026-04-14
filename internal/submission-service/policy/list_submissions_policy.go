package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ListSubmissionsParams struct {
	HackathonID     uuid.UUID
	TargetOwnerKind string
	TargetOwnerID   uuid.UUID
	ActorOwnerKind  string
	ActorOwnerID    uuid.UUID
}

type ListSubmissionsContext struct {
	authenticated   bool
	actorUserID     uuid.UUID
	actorRoles      []string
	hackathonStage  string
	targetOwnerKind string
	targetOwnerID   uuid.UUID
	actorOwnerKind  string
	actorOwnerID    uuid.UUID
}

func NewListSubmissionsContext() *ListSubmissionsContext {
	return &ListSubmissionsContext{}
}

func (c *ListSubmissionsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *ListSubmissionsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *ListSubmissionsContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *ListSubmissionsContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *ListSubmissionsContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *ListSubmissionsContext) SetTargetOwner(kind string, id uuid.UUID) {
	c.targetOwnerKind = kind
	c.targetOwnerID = id
}

func (c *ListSubmissionsContext) SetActorOwner(kind string, id uuid.UUID) {
	c.actorOwnerKind = kind
	c.actorOwnerID = id
}

func (c *ListSubmissionsContext) IsStaff() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleOwner || role == domain.RoleOrganizer || role == domain.RoleMentor || role == domain.RoleJudge {
			return true
		}
	}
	return false
}

func (c *ListSubmissionsContext) IsOwnSubmissions() bool {
	return c.targetOwnerKind == c.actorOwnerKind && c.targetOwnerID == c.actorOwnerID
}

func (c *ListSubmissionsContext) IsReadWindowStage() bool {
	return c.hackathonStage == domain.HackathonStageRunning ||
		c.hackathonStage == domain.HackathonStageJudging ||
		c.hackathonStage == domain.HackathonStageFinished
}

type ListSubmissionsPolicy struct{}

func NewListSubmissionsPolicy() *ListSubmissionsPolicy {
	return &ListSubmissionsPolicy{}
}

func (p *ListSubmissionsPolicy) Action() policy.Action {
	return ActionListSubmissions
}

func (p *ListSubmissionsPolicy) LoadContext(ctx context.Context, params ListSubmissionsParams) (*ListSubmissionsContext, error) {
	pctx := NewListSubmissionsContext()
	pctx.SetTargetOwner(params.TargetOwnerKind, params.TargetOwnerID)
	pctx.SetActorOwner(params.ActorOwnerKind, params.ActorOwnerID)
	return pctx, nil
}

func (p *ListSubmissionsPolicy) Check(ctx context.Context, pctx *ListSubmissionsContext) *policy.Decision {
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

	if !pctx.IsOwnSubmissions() && !pctx.IsStaff() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only staff can list other participants' submissions",
		})
		return decision
	}

	decision.Allow()
	return decision
}
