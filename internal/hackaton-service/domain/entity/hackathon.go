package entity

import (
	"time"

	"github.com/google/uuid"
)

type Hackathon struct {
	ID               uuid.UUID
	Name             string
	ShortDescription string
	Description      string

	LocationOnline  bool
	LocationCity    string
	LocationCountry string
	LocationVenue   string

	StartsAt             *time.Time
	EndsAt               *time.Time
	RegistrationOpensAt  *time.Time
	RegistrationClosesAt *time.Time
	SubmissionsOpensAt   *time.Time
	SubmissionsClosesAt  *time.Time
	JudgingEndsAt        *time.Time

	Stage string
	State string

	PublishedAt       *time.Time
	ResultPublishedAt *time.Time

	TeamSizeMax int32

	AllowIndividual bool
	AllowTeam       bool

	Task   string
	Result string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type HackathonLink struct {
	ID          uuid.UUID
	HackathonID uuid.UUID
	Title       string
	URL         string
}
