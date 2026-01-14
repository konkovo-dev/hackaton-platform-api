package entity

import (
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain"
	"github.com/google/uuid"
)

type Contact struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Type       string
	Value      string
	Visibility domain.VisibilityLevel
}
