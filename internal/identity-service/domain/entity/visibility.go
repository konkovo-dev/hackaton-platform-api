package entity

import (
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain"
	"github.com/google/uuid"
)

type Visibility struct {
	UserID             uuid.UUID
	SkillsVisibility   domain.VisibilityLevel
	ContactsVisibility domain.VisibilityLevel
}

func DefaultVisibility(id uuid.UUID) *Visibility {
	return &Visibility{
		UserID:             id,
		SkillsVisibility:   domain.VisibilityLevelPublic,
		ContactsVisibility: domain.VisibilityLevelPublic,
	}
}
