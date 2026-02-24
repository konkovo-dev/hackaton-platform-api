package entity

import (
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	ID                    uuid.UUID
	HackathonID           uuid.UUID
	OwnerKind             string
	OwnerID               uuid.UUID
	Status                string
	AssignedMentorUserID  *uuid.UUID
	CreatedAt             time.Time
	UpdatedAt             time.Time
	ClosedAt              *time.Time
}
