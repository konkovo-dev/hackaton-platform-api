package policy

import (
	"time"

	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type HackathonPolicyContext struct {
	authenticated bool
	actorUserID   uuid.UUID

	roles             []string
	participationKind string
	teamID            string

	hackathonID       uuid.UUID
	stage             string
	state             string
	publishedAt       *time.Time
	resultPublishedAt *time.Time
}

func NewHackathonPolicyContext() *HackathonPolicyContext {
	return &HackathonPolicyContext{}
}

func (c *HackathonPolicyContext) SetAuthenticated(auth bool) {
	c.authenticated = auth
}

func (c *HackathonPolicyContext) IsAuthenticated() bool {
	return c.authenticated
}

func (c *HackathonPolicyContext) SetActorUserID(id uuid.UUID) {
	c.actorUserID = id
}

func (c *HackathonPolicyContext) ActorUserID() uuid.UUID {
	return c.actorUserID
}

func (c *HackathonPolicyContext) SetRoles(roles []string) {
	c.roles = roles
}

func (c *HackathonPolicyContext) Roles() []string {
	return c.roles
}

func (c *HackathonPolicyContext) HasRole(role string) bool {
	for _, r := range c.roles {
		if r == role {
			return true
		}
	}
	return false
}

func (c *HackathonPolicyContext) SetParticipationKind(kind string) {
	c.participationKind = kind
}

func (c *HackathonPolicyContext) ParticipationKind() string {
	return c.participationKind
}

func (c *HackathonPolicyContext) SetHackathonID(id uuid.UUID) {
	c.hackathonID = id
}

func (c *HackathonPolicyContext) HackathonID() uuid.UUID {
	return c.hackathonID
}

func (c *HackathonPolicyContext) SetStage(stage string) {
	c.stage = stage
}

func (c *HackathonPolicyContext) Stage() string {
	return c.stage
}

func (c *HackathonPolicyContext) SetState(state string) {
	c.state = state
}

func (c *HackathonPolicyContext) State() string {
	return c.state
}

func (c *HackathonPolicyContext) SetPublishedAt(t *time.Time) {
	c.publishedAt = t
}

func (c *HackathonPolicyContext) PublishedAt() *time.Time {
	return c.publishedAt
}

func (c *HackathonPolicyContext) SetResultPublishedAt(t *time.Time) {
	c.resultPublishedAt = t
}

func (c *HackathonPolicyContext) ResultPublishedAt() *time.Time {
	return c.resultPublishedAt
}

func (c *HackathonPolicyContext) SetTeamID(teamID string) {
	c.teamID = teamID
}

func (c *HackathonPolicyContext) TeamID() string {
	return c.teamID
}

var _ policy.PolicyContext = (*HackathonPolicyContext)(nil)
