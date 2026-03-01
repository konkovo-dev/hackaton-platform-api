package entity

import (
	"time"

	"github.com/google/uuid"
)

type Submission struct {
	ID              uuid.UUID
	HackathonID     uuid.UUID
	OwnerKind       string
	OwnerID         uuid.UUID
	CreatedByUserID uuid.UUID
	Title           string
	Description     string
	IsFinal         bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
