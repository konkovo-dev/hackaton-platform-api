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

func (s *MeService) GetMe(ctx context.Context, req *identityv1.GetMeRequest) (*identityv1.GetMeResponse, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		s.logger.WarnContext(ctx, "get_me: no user_id in context")
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.WarnContext(ctx, "get_me: invalid user_id", slog.String("user_id", userID))
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	result, err := s.meService.GetMe(ctx, me.GetMeIn{
		UserID: userUUID,
	})
	if err != nil {
		return nil, s.handleError(ctx, err, "get_me")
	}

	s.logger.InfoContext(ctx, "get_me: success", slog.String("user_id", userID))
	return mappers.GetMeOutToResponse(result), nil
}
