package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type GetMyAssignmentsParams struct {
	HackathonID uuid.UUID
}

type GetMyAssignmentsContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
}

func NewGetMyAssignmentsContext() *GetMyAssignmentsContext {
	return &GetMyAssignmentsContext{}
}

func (c *GetMyAssignmentsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *GetMyAssignmentsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *GetMyAssignmentsContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *GetMyAssignmentsContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *GetMyAssignmentsContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *GetMyAssignmentsContext) IsJudge() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleJudge {
			return true
		}
	}
	return false
}

func (c *GetMyAssignmentsContext) IsJudgingOrLater() bool {
	return c.hackathonStage == domain.HackathonStageJudging ||
		c.hackathonStage == domain.HackathonStageFinished
}

type GetMyAssignmentsPolicy struct{}

func NewGetMyAssignmentsPolicy() *GetMyAssignmentsPolicy {
	return &GetMyAssignmentsPolicy{}
}

func (p *GetMyAssignmentsPolicy) Action() policy.Action {
	return ActionGetMyAssignments
}

func (p *GetMyAssignmentsPolicy) LoadContext(ctx context.Context, params GetMyAssignmentsParams) (*GetMyAssignmentsContext, error) {
	return NewGetMyAssignmentsContext(), nil
}

func (p *GetMyAssignmentsPolicy) Check(ctx context.Context, pctx *GetMyAssignmentsContext) *policy.Decision {
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
			Message: "assignments can only be viewed from JUDGING stage onwards",
		})
		return decision
	}

	if !pctx.IsJudge() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only judges can view their assignments",
		})
		return decision
	}

	decision.Allow()
	return decision
}
