package entity

import (
	"time"

	"github.com/google/uuid"
)

type Participation struct {
	HackathonID uuid.UUID
	UserID      uuid.UUID
	Status      string
	TeamID      *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
