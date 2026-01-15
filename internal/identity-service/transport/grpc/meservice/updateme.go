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

func (s *MeService) UpdateMe(ctx context.Context, req *identityv1.UpdateMeRequest) (*identityv1.UpdateMeResponse, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		s.logger.WarnContext(ctx, "update_me: no user_id in context")
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.WarnContext(ctx, "update_me: invalid user_id", slog.String("user_id", userID))
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	result, err := s.meService.UpdateMe(ctx, me.UpdateMeIn{
		UserID:    userUUID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		AvatarURL: req.AvatarUrl,
		Timezone:  req.Timezone,
	})
	if err != nil {
		return nil, s.handleError(ctx, err, "update_me")
	}

	s.logger.InfoContext(ctx, "update_me: success", slog.String("user_id", userID))
	return mappers.UpdateMeOutToResponse(result), nil
}
