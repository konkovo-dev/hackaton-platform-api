package entity

import "github.com/google/uuid"

type CatalogSkill struct {
	ID   uuid.UUID
	Name string
}

type CustomSkill struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string
}
