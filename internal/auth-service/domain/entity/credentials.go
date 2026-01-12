package entity

import (
	"time"

	"github.com/google/uuid"
)

type Credentials struct {
	UserID       uuid.UUID
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
