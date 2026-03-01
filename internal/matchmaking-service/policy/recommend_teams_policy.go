package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type RecommendTeamsParams struct {
	HackathonID uuid.UUID
	UserID      uuid.UUID
}

type RecommendTeamsContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
	requestUserID  uuid.UUID
}

func NewRecommendTeamsContext(requestUserID uuid.UUID) *RecommendTeamsContext {
	return &RecommendTeamsContext{
		requestUserID: requestUserID,
	}
}

func (c *RecommendTeamsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *RecommendTeamsContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *RecommendTeamsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *RecommendTeamsContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *RecommendTeamsContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *RecommendTeamsContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *RecommendTeamsContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *RecommendTeamsContext) ActorRoles() []string {
	return c.actorRoles
}

func (c *RecommendTeamsContext) IsParticipant() bool {
	for _, role := range c.actorRoles {
		if role == "participant" {
			return true
		}
	}
	return false
}

type RecommendTeamsPolicy struct{}

func NewRecommendTeamsPolicy() *RecommendTeamsPolicy {
	return &RecommendTeamsPolicy{}
}

func (p *RecommendTeamsPolicy) Action() policy.Action {
	return ActionRecommendTeams
}

func (p *RecommendTeamsPolicy) LoadContext(ctx context.Context, params RecommendTeamsParams) (*RecommendTeamsContext, error) {
	return NewRecommendTeamsContext(params.UserID), nil
}

func (p *RecommendTeamsPolicy) Check(ctx context.Context, pctx *RecommendTeamsContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if pctx.ActorUserID() != pctx.requestUserID {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "users can only get recommendations for themselves",
		})
		return decision
	}

	if pctx.HackathonStage() != "REGISTRATION" && pctx.HackathonStage() != "RUNNING" {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeStageRule,
			Message: "matchmaking is only available during REGISTRATION and RUNNING stages",
		})
		return decision
	}

	if !pctx.IsParticipant() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only participants can get team recommendations",
		})
		return decision
	}

	decision.Allow()
	return decision
}
