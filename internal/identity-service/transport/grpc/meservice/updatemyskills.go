package meservice

import (
	"context"
	"log/slog"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/mappers"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/me"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *MeService) UpdateMySkills(ctx context.Context, req *identityv1.UpdateMySkillsRequest) (*identityv1.UpdateMySkillsResponse, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		s.logger.WarnContext(ctx, "update_my_skills: no user_id in context")
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.WarnContext(ctx, "update_my_skills: invalid user_id", slog.String("user_id", userID))
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	catalogSkillIDs := make([]uuid.UUID, 0, len(req.CatalogSkillIds))
	for _, idStr := range req.CatalogSkillIds {
		id, err := uuid.Parse(idStr)
		if err != nil {
			s.logger.WarnContext(ctx, "update_my_skills: invalid catalog_skill_id", slog.String("id", idStr))
			return nil, status.Error(codes.InvalidArgument, "invalid catalog_skill_id format")
		}
		catalogSkillIDs = append(catalogSkillIDs, id)
	}

	result, err := s.meService.UpdateMySkills(ctx, me.UpdateMySkillsIn{
		UserID:           userUUID,
		CatalogSkillIDs:  catalogSkillIDs,
		CustomSkills:     req.UserSkills,
		SkillsVisibility: mappers.ProtoVisibilityLevelToDomain(req.SkillsVisibility),
	})
	if err != nil {
		return nil, s.handleError(ctx, err, "update_my_skills")
	}

	s.logger.InfoContext(ctx, "update_my_skills: success", slog.String("user_id", userID))
	return mappers.UpdateMySkillsOutToResponse(result), nil
}
