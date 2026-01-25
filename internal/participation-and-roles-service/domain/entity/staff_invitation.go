package entity

import (
	"time"

	"github.com/google/uuid"
)

type StaffInvitation struct {
	ID              uuid.UUID
	HackathonID     uuid.UUID
	TargetUserID    uuid.UUID
	RequestedRole   string
	CreatedByUserID uuid.UUID
	Message         string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ExpiresAt       *time.Time
}
