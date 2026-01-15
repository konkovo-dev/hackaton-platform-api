package skills

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/repository/postgres"
	"github.com/google/uuid"
)

type SkillRepository interface {
	ListSkillCatalog(ctx context.Context, params postgres.ListSkillCatalogParams) ([]*postgres.ListSkillCatalogResult, bool, error)
}

type SkillCatalogResult struct {
	ID   uuid.UUID
	Name string
}
