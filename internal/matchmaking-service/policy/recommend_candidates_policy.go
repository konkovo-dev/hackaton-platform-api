package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type RecommendCandidatesParams struct {
	HackathonID uuid.UUID
	UserID      uuid.UUID
	VacancyID   uuid.UUID
}

type RecommendCandidatesContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	actorRoles     []string
	requestUserID  uuid.UUID
}

func NewRecommendCandidatesContext(requestUserID uuid.UUID) *RecommendCandidatesContext {
	return &RecommendCandidatesContext{
		requestUserID: requestUserID,
	}
}

func (c *RecommendCandidatesContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *RecommendCandidatesContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *RecommendCandidatesContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *RecommendCandidatesContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *RecommendCandidatesContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *RecommendCandidatesContext) HackathonStage() string {
	return c.hackathonStage
}

func (c *RecommendCandidatesContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *RecommendCandidatesContext) ActorRoles() []string {
	return c.actorRoles
}

func (c *RecommendCandidatesContext) IsParticipant() bool {
	for _, role := range c.actorRoles {
		if role == "participant" {
			return true
		}
	}
	return false
}

type RecommendCandidatesPolicy struct{}

func NewRecommendCandidatesPolicy() *RecommendCandidatesPolicy {
	return &RecommendCandidatesPolicy{}
}

func (p *RecommendCandidatesPolicy) Action() policy.Action {
	return ActionRecommendCandidates
}

func (p *RecommendCandidatesPolicy) LoadContext(ctx context.Context, params RecommendCandidatesParams) (*RecommendCandidatesContext, error) {
	return NewRecommendCandidatesContext(params.UserID), nil
}

func (p *RecommendCandidatesPolicy) Check(ctx context.Context, pctx *RecommendCandidatesContext) *policy.Decision {
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
			Message: "only participants can get candidate recommendations",
		})
		return decision
	}

	decision.Allow()
	return decision
}
