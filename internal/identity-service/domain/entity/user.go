package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Username  string
	FirstName string
	LastName  string
	AvatarURL string
	Timezone  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
