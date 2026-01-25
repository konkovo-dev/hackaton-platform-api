package policy

import (
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type AuthPolicyContext struct {
	authenticated bool
	actorUserID   uuid.UUID
}

func NewAuthPolicyContext() *AuthPolicyContext {
	return &AuthPolicyContext{
		authenticated: false,
		actorUserID:   uuid.Nil,
	}
}

func (c *AuthPolicyContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *AuthPolicyContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

var _ policy.PolicyContext = (*AuthPolicyContext)(nil)
