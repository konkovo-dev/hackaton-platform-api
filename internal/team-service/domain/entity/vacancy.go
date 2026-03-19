package entity

import (
	"time"

	"github.com/google/uuid"
)

type Vacancy struct {
	ID               uuid.UUID
	TeamID           uuid.UUID
	Description      string
	DesiredRoleIDs   []uuid.UUID
	DesiredSkillIDs  []uuid.UUID
	SlotsTotal       int64
	SlotsOpen        int64
	IsSystem         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
