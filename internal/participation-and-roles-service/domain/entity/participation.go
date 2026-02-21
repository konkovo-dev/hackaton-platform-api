package entity

import (
	"time"

	"github.com/google/uuid"
)

type Participation struct {
	HackathonID    uuid.UUID
	UserID         uuid.UUID
	Status         string
	TeamID         *uuid.UUID
	WishedRoles    []*TeamRole
	MotivationText string
	RegisteredAt   time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TeamRole struct {
	ID   uuid.UUID
	Name string
}
