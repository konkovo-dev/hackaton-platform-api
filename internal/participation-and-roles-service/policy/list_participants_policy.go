package policy

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ListParticipantsParams struct {
	HackathonID uuid.UUID
}

type ListParticipantsContext struct {
	authenticated bool
	actorUserID   uuid.UUID

	actorRoles          []string
	participationStatus string
}

func NewListParticipantsContext() *ListParticipantsContext {
	return &ListParticipantsContext{}
}

func (c *ListParticipantsContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *ListParticipantsContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *ListParticipantsContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *ListParticipantsContext) SetActorUserID(id uuid.UUID) {
	c.actorUserID = id
}

func (c *ListParticipantsContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *ListParticipantsContext) IsStaff() bool {
	for _, role := range c.actorRoles {
		if role == string(domain.RoleOwner) ||
			role == string(domain.RoleOrganizer) ||
			role == string(domain.RoleMentor) {
			return true
		}
	}
	return false
}

func (c *ListParticipantsContext) ActorRoles() []string {
	return c.actorRoles
}

func (c *ListParticipantsContext) SetParticipationStatus(status string) {
	c.participationStatus = status
}

func (c *ListParticipantsContext) IsParticipant() bool {
	return c.participationStatus != "" && c.participationStatus != "none"
}

type ListParticipantsRepositories interface {
	GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error)
	GetParticipationStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error)
}

type ListParticipantsPolicy struct {
	repos ListParticipantsRepositories
}

func NewListParticipantsPolicy(repos ListParticipantsRepositories) *ListParticipantsPolicy {
	return &ListParticipantsPolicy{
		repos: repos,
	}
}

func (p *ListParticipantsPolicy) LoadContext(ctx context.Context, params ListParticipantsParams) (*ListParticipantsContext, error) {
	pctx := NewListParticipantsContext()

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

func (p *ListParticipantsPolicy) Check(ctx context.Context, pctx *ListParticipantsContext) *policy.Decision {
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
			Message: "only staff or participants can list participants",
		})
		return decision
	}

	decision.Allow()
	return decision
}
