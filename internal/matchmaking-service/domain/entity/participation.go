package entity

import (
	"time"

	"github.com/google/uuid"
)

type Participation struct {
	HackathonID    uuid.UUID
	UserID         uuid.UUID
	Status         string
	WishedRoleIDs  []uuid.UUID
	MotivationText string
	TeamID         *uuid.UUID
	UpdatedAt      time.Time
}
