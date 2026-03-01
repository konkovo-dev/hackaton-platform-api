package entity

import (
	"time"

	"github.com/google/uuid"
)

type Vacancy struct {
	VacancyID        uuid.UUID
	TeamID           uuid.UUID
	HackathonID      uuid.UUID
	Description      string
	DesiredRoleIDs   []uuid.UUID
	DesiredSkillIDs  []uuid.UUID
	SlotsOpen        int32
	UpdatedAt        time.Time
}
