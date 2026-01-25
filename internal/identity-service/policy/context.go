package policy

import (
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type IdentityPolicyContext struct {
	authenticated bool
	actorUserID   uuid.UUID

	targetUserID uuid.UUID
	isMe         bool

	skillsVisibility   bool
	contactsVisibility bool
}

func NewIdentityPolicyContext() *IdentityPolicyContext {
	return &IdentityPolicyContext{
		authenticated: false,
		actorUserID:   uuid.Nil,
		targetUserID:  uuid.Nil,
		isMe:          false,
	}
}

func (c *IdentityPolicyContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *IdentityPolicyContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *IdentityPolicyContext) SetAuthenticated(authenticated bool) {
	c.authenticated = authenticated
}

func (c *IdentityPolicyContext) SetActorUserID(userID uuid.UUID) {
	c.actorUserID = userID
}

func (c *IdentityPolicyContext) SetTargetUserID(userID uuid.UUID) {
	c.targetUserID = userID
}

func (c *IdentityPolicyContext) SetIsMe(isMe bool) {
	c.isMe = isMe
}

func (c *IdentityPolicyContext) SetVisibility(skillsVisibility, contactsVisibility domain.VisibilityLevel) {
	c.skillsVisibility = skillsVisibility == domain.VisibilityLevelPublic
	c.contactsVisibility = contactsVisibility == domain.VisibilityLevelPublic
}

func (c *IdentityPolicyContext) IsMe() bool {
	return c.isMe
}

func (c *IdentityPolicyContext) TargetUserID() uuid.UUID {
	return c.targetUserID
}

func (c *IdentityPolicyContext) SkillsVisibility() bool {
	return c.skillsVisibility
}

func (c *IdentityPolicyContext) ContactsVisibility() bool {
	return c.contactsVisibility
}

var _ policy.PolicyContext = (*IdentityPolicyContext)(nil)
