package policy

import (
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type StaffPolicyContext struct {
	authenticated bool
	actorUserID   uuid.UUID

	roles []string

	hackathonID uuid.UUID
}

func NewStaffPolicyContext() *StaffPolicyContext {
	return &StaffPolicyContext{}
}

func (c *StaffPolicyContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *StaffPolicyContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *StaffPolicyContext) SetActorUserID(id uuid.UUID) {
	c.actorUserID = id
}

func (c *StaffPolicyContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *StaffPolicyContext) SetRoles(roles []string) {
	c.roles = roles
}

func (c *StaffPolicyContext) Roles() []string {
	return c.roles
}

func (c *StaffPolicyContext) HasRole(role string) bool {
	for _, r := range c.roles {
		if r == role {
			return true
		}
	}
	return false
}

func (c *StaffPolicyContext) SetHackathonID(id uuid.UUID) {
	c.hackathonID = id
}

func (c *StaffPolicyContext) HackathonID() uuid.UUID {
	return c.hackathonID
}

var _ policy.PolicyContext = (*StaffPolicyContext)(nil)
