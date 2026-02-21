package entity

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	ID          uuid.UUID
	HackathonID uuid.UUID
	Name        string
	Description string
	IsJoinable  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
