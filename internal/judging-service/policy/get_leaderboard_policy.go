package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type GetLeaderboardParams struct {
	HackathonID uuid.UUID
}

type GetLeaderboardContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
}

func NewGetLeaderboardContext() *GetLeaderboardContext {
	return &GetLeaderboardContext{}
}

func (c *GetLeaderboardContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *GetLeaderboardContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *GetLeaderboardContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *GetLeaderboardContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *GetLeaderboardContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *GetLeaderboardContext) IsStaffOrJudge() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleOwner || role == domain.RoleOrganizer || role == domain.RoleJudge {
			return true
		}
	}
	return false
}

func (c *GetLeaderboardContext) IsJudgingOrLater() bool {
	return c.hackathonStage == domain.HackathonStageJudging ||
		c.hackathonStage == domain.HackathonStageFinished
}

type GetLeaderboardPolicy struct{}

func NewGetLeaderboardPolicy() *GetLeaderboardPolicy {
	return &GetLeaderboardPolicy{}
}

func (p *GetLeaderboardPolicy) Action() policy.Action {
	return ActionGetLeaderboard
}

func (p *GetLeaderboardPolicy) LoadContext(ctx context.Context, params GetLeaderboardParams) (*GetLeaderboardContext, error) {
	return NewGetLeaderboardContext(), nil
}

func (p *GetLeaderboardPolicy) Check(ctx context.Context, pctx *GetLeaderboardContext) *policy.Decision {
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
			Message: "leaderboard can only be viewed from JUDGING stage onwards",
		})
		return decision
	}

	if !pctx.IsStaffOrJudge() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only organizers and judges can view leaderboard",
		})
		return decision
	}

	decision.Allow()
	return decision
}
