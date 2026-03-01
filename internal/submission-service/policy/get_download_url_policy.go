package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type GetDownloadURLParams struct {
	HackathonID uuid.UUID
	SubmissionID uuid.UUID
	OwnerKind   string
	OwnerID     uuid.UUID
	ActorOwnerKind string
	ActorOwnerID   uuid.UUID
}

type GetDownloadURLContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	actorRoles     []string
	ownerKind      string
	ownerID        uuid.UUID
	actorOwnerKind string
	actorOwnerID   uuid.UUID
}

func NewGetDownloadURLContext() *GetDownloadURLContext {
	return &GetDownloadURLContext{}
}

func (c *GetDownloadURLContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *GetDownloadURLContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *GetDownloadURLContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *GetDownloadURLContext) SetActorRoles(roles []string) {
	c.actorRoles = roles
}

func (c *GetDownloadURLContext) SetOwner(kind string, id uuid.UUID) {
	c.ownerKind = kind
	c.ownerID = id
}

func (c *GetDownloadURLContext) SetActorOwner(kind string, id uuid.UUID) {
	c.actorOwnerKind = kind
	c.actorOwnerID = id
}

func (c *GetDownloadURLContext) IsStaff() bool {
	for _, role := range c.actorRoles {
		if role == domain.RoleOwner || role == domain.RoleOrganizer || role == domain.RoleMentor || role == domain.RoleJudge {
			return true
		}
	}
	return false
}

func (c *GetDownloadURLContext) IsOwner() bool {
	return c.ownerKind == c.actorOwnerKind && c.ownerID == c.actorOwnerID
}

type GetDownloadURLPolicy struct{}

func NewGetDownloadURLPolicy() *GetDownloadURLPolicy {
	return &GetDownloadURLPolicy{}
}

func (p *GetDownloadURLPolicy) Action() policy.Action {
	return ActionGetSubmissionFileDownloadURL
}

func (p *GetDownloadURLPolicy) LoadContext(ctx context.Context, params GetDownloadURLParams) (*GetDownloadURLContext, error) {
	pctx := NewGetDownloadURLContext()
	pctx.SetOwner(params.OwnerKind, params.OwnerID)
	pctx.SetActorOwner(params.ActorOwnerKind, params.ActorOwnerID)
	return pctx, nil
}

func (p *GetDownloadURLPolicy) Check(ctx context.Context, pctx *GetDownloadURLContext) *policy.Decision {
	decision := policy.NewDecision()

	if !pctx.IsAuthenticated() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "user must be authenticated",
		})
		return decision
	}

	if !pctx.IsOwner() && !pctx.IsStaff() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only owner or staff can download submission files",
		})
		return decision
	}

	decision.Allow()
	return decision
}
