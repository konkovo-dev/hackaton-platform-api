package entity

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	TeamID      uuid.UUID
	HackathonID uuid.UUID
	Name        string
	Description string
	IsJoinable  bool
	UpdatedAt   time.Time
}
