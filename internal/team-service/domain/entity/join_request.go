package entity

import (
	"time"

	"github.com/google/uuid"
)

type JoinRequest struct {
	ID              uuid.UUID
	HackathonID     uuid.UUID
	TeamID          uuid.UUID
	VacancyID       uuid.UUID
	RequesterUserID uuid.UUID
	Message         string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ExpiresAt       *time.Time
}
