package policy

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type FileOperationParams struct {
	HackathonID uuid.UUID
	SubmissionID uuid.UUID
	OwnerKind   string
	OwnerID     uuid.UUID
	ActorOwnerKind string
	ActorOwnerID   uuid.UUID
}

type FileOperationContext struct {
	authenticated  bool
	actorUserID    uuid.UUID
	hackathonStage string
	ownerKind      string
	ownerID        uuid.UUID
	actorOwnerKind string
	actorOwnerID   uuid.UUID
}

func NewFileOperationContext() *FileOperationContext {
	return &FileOperationContext{}
}

func (c *FileOperationContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *FileOperationContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *FileOperationContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *FileOperationContext) SetHackathonStage(stage string) {
	c.hackathonStage = stage
}

func (c *FileOperationContext) SetOwner(kind string, id uuid.UUID) {
	c.ownerKind = kind
	c.ownerID = id
}

func (c *FileOperationContext) SetActorOwner(kind string, id uuid.UUID) {
	c.actorOwnerKind = kind
	c.actorOwnerID = id
}

func (c *FileOperationContext) IsOwner() bool {
	return c.ownerKind == c.actorOwnerKind && c.ownerID == c.actorOwnerID
}

type FileOperationPolicy struct {
	action policy.Action
}

func NewFileOperationPolicy(action policy.Action) *FileOperationPolicy {
	return &FileOperationPolicy{action: action}
}

func (p *FileOperationPolicy) Action() policy.Action {
	return p.action
}

func (p *FileOperationPolicy) LoadContext(ctx context.Context, params FileOperationParams) (*FileOperationContext, error) {
	pctx := NewFileOperationContext()
	pctx.SetOwner(params.OwnerKind, params.OwnerID)
	pctx.SetActorOwner(params.ActorOwnerKind, params.ActorOwnerID)
	return pctx, nil
}

func (p *FileOperationPolicy) Check(ctx context.Context, pctx *FileOperationContext) *policy.Decision {
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
			Message: "file operations are only available during RUNNING stage",
		})
		return decision
	}

	if !pctx.IsOwner() {
		decision.Deny(policy.Violation{
			Code:    policy.ViolationCodeForbidden,
			Message: "only the owner can perform file operations",
		})
		return decision
	}

	decision.Allow()
	return decision
}
