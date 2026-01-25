package entity

import (
	"time"

	"github.com/google/uuid"
)

type HackathonAnnouncement struct {
	ID            uuid.UUID
	HackathonID   uuid.UUID
	Title         string
	Body          string
	CreatedByUser uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}
