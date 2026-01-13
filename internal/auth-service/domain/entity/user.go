package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Username  string
	Email     string
	FirstName string
	LastName  string
	Timezone  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
