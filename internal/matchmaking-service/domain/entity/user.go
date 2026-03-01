package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID           uuid.UUID
	Username         string
	AvatarURL        string
	CatalogSkillIDs  []uuid.UUID
	CustomSkillNames []string
	UpdatedAt        time.Time
}
