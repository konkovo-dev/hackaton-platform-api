package entity

import (
	"time"

	"github.com/google/uuid"
)

type Membership struct {
	TeamID             uuid.UUID
	UserID             uuid.UUID
	IsCaptain          bool
	AssignedVacancyID  *uuid.UUID
	JoinedAt           time.Time
}
