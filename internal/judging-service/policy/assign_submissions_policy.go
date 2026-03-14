package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type AssignSubmissionsParams struct {
	HackathonID uuid.UUID
}

type AssignSubmissionsContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
}

func NewAssignSubmissionsContext() *AssignSubmissionsContext {
	return &AssignSubmissionsContext{}
}

func (c *AssignSubmissionsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *AssignSubmissionsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *AssignSubmissionsContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *AssignSubmissionsContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *AssignSubmissionsContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *AssignSubmissionsContext) IsOrganizer() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleOwner || role == domain.RoleOrganizer {
			return true
		}
	}
	return false
}

type AssignSubmissionsPolicy struct{}

func NewAssignSubmissionsPolicy() *AssignSubmissionsPolicy {
	return &AssignSubmissionsPolicy{}
}

func (p *AssignSubmissionsPolicy) Action() policy.Action {
	return ActionAssignSubmissionsToJudges
}

func (p *AssignSubmissionsPolicy) LoadContext(ctx context.Context, params AssignSubmissionsParams) (*AssignSubmissionsContext, error) {
	return NewAssignSubmissionsContext(), nil
}

func (p *AssignSubmissionsPolicy) Check(ctx context.Context, pctx *AssignSubmissionsContext) *policy.Decision {
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
			Message: "assignments can only be created during JUDGING stage",
		})
		return decision
	}

	if !pctx.IsOrganizer() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only organizers can assign submissions to judges",
		})
		return decision
	}

	decision.Allow()
	return decision
}
