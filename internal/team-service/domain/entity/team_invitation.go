package entity

import (
	"time"

	"github.com/google/uuid"
)

type TeamInvitation struct {
	ID                uuid.UUID
	HackathonID       uuid.UUID
	TeamID            uuid.UUID
	VacancyID         uuid.UUID
	TargetUserID      uuid.UUID
	CreatedByUserID   uuid.UUID
	Message           string
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ExpiresAt         *time.Time
}
