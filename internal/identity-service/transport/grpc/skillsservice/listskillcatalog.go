package skillsservice

import (
	"context"
	"log/slog"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/mappers"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/skills"
)

func (s *SkillsService) ListSkillCatalog(ctx context.Context, req *identityv1.ListSkillCatalogRequest) (*identityv1.ListSkillCatalogResponse, error) {
	result, err := s.skillsService.ListSkillCatalog(ctx, skills.ListSkillCatalogIn{
		Query: req.Query,
	})
	if err != nil {
		return nil, s.handleError(ctx, err, "list_skill_catalog")
	}

	s.logger.InfoContext(ctx, "list_skill_catalog: success", slog.Int("count", len(result.Skills)))
	return mappers.ListSkillCatalogOutToResponse(result), nil
}
