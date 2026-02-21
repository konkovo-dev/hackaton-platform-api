package policy

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type GetUserParticipationParams struct {
	HackathonID uuid.UUID
	TargetUserID uuid.UUID
}

type GetUserParticipationContext struct {
	authenticated bool
	actorUserID   uuid.UUID
	targetUserID  uuid.UUID

	actorRoles          []string
	participationStatus string
}

func NewGetUserParticipationContext() *GetUserParticipationContext {
	return &GetUserParticipationContext{}
}

func (c *GetUserParticipationContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *GetUserParticipationContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *GetUserParticipationContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *GetUserParticipationContext) SetActorUserID(id uuid.UUID) {
	c.actorUserID = id
}

func (c *GetUserParticipationContext) SetTargetUserID(id uuid.UUID) {
	c.targetUserID = id
}

func (c *GetUserParticipationContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *GetUserParticipationContext) IsStaff() bool {
	for _, role := range c.actorRoles {
		if role == string(domain.RoleOwner) ||
			role == string(domain.RoleOrganizer) ||
			role == string(domain.RoleMentor) {
			return true
		}
	}
	return false
}

func (c *GetUserParticipationContext) ActorRoles() []string {
	return c.actorRoles
}

func (c *GetUserParticipationContext) TargetUserID() uuid.UUID {
	return c.targetUserID
}

func (c *GetUserParticipationContext) SetParticipationStatus(status string) {
	c.participationStatus = status
}

func (c *GetUserParticipationContext) IsParticipant() bool {
	return c.participationStatus != "" && c.participationStatus != "none"
}

type GetUserParticipationRepositories interface {
	GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error)
	GetParticipationStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error)
}

type GetUserParticipationPolicy struct {
	repos GetUserParticipationRepositories
}

func NewGetUserParticipationPolicy(repos GetUserParticipationRepositories) *GetUserParticipationPolicy {
	return &GetUserParticipationPolicy{
		repos: repos,
	}
}

func (p *GetUserParticipationPolicy) LoadContext(ctx context.Context, params GetUserParticipationParams) (*GetUserParticipationContext, error) {
	pctx := NewGetUserParticipationContext()

	userID, ok := auth.GetUserID(ctx)
	if !ok {
		pctx.SetAuthenticated(false)
		return pctx, nil
	}

	pctx.SetAuthenticated(true)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user id: %w", err)
	}

	pctx.SetActorUserID(userUUID)
	pctx.SetTargetUserID(params.TargetUserID)

	roles, err := p.repos.GetRoleStrings(ctx, params.HackathonID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get actor roles: %w", err)
	}

	pctx.SetActorRoles(roles)

	participationStatus, err := p.repos.GetParticipationStatus(ctx, params.HackathonID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participation status: %w", err)
	}

	pctx.SetParticipationStatus(participationStatus)

	return pctx, nil
}

func (p *GetUserParticipationPolicy) Check(ctx context.Context, pctx *GetUserParticipationContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if !pctx.IsStaff() && !pctx.IsParticipant() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only staff or participants can view user participations",
		})
		return decision
	}

	decision.Allow()
	return decision
}
