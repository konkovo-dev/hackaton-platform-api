package entity

import (
	"time"

	"github.com/google/uuid"
)

type StaffRole struct {
	HackathonID uuid.UUID
	UserID      uuid.UUID
	Role        string
	CreatedAt   time.Time
}
