package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type CreateSubmissionParams struct {
	HackathonID uuid.UUID
}

type CreateSubmissionContext struct {
	authenticated       bool
	actorUserID         uuid.UUID
	hackathonStage      string
	participationStatus string
	actorRoles          []string
}

func NewCreateSubmissionContext() *CreateSubmissionContext {
	return &CreateSubmissionContext{}
}

func (c *CreateSubmissionContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *CreateSubmissionContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *CreateSubmissionContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *CreateSubmissionContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *CreateSubmissionContext) SetParticipationStatus(status string) {
	c.participationStatus = status
}

func (c *CreateSubmissionContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *CreateSubmissionContext) IsStaff() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleOwner || role == domain.RoleOrganizer || role == domain.RoleMentor || role == domain.RoleJudge {
			return true
		}
	}
	return false
}

func (c *CreateSubmissionContext) IsParticipant() bool {
	return c.participationStatus == "individual_active" ||
		c.participationStatus == "team_member" ||
		c.participationStatus == "team_captain"
}

type CreateSubmissionPolicy struct{}

func NewCreateSubmissionPolicy() *CreateSubmissionPolicy {
	return &CreateSubmissionPolicy{}
}

func (p *CreateSubmissionPolicy) Action() policy.Action {
	return ActionCreateSubmission
}

func (p *CreateSubmissionPolicy) LoadContext(ctx context.Context, params CreateSubmissionParams) (*CreateSubmissionContext, error) {
	return NewCreateSubmissionContext(), nil
}

func (p *CreateSubmissionPolicy) Check(ctx context.Context, pctx *CreateSubmissionContext) *policy.Decision {
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
			Message: "submissions can only be created during RUNNING stage",
		})
		return decision
	}

	if !pctx.IsParticipant() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only active participants can create submissions",
		})
		return decision
	}

	decision.Allow()
	return decision
}
