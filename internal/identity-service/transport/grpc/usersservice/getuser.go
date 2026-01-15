package usersservice

import (
	"context"
	"log/slog"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/mappers"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/users"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UsersService) GetUser(ctx context.Context, req *identityv1.GetUserRequest) (*identityv1.GetUserResponse, error) {
	if req.UserId == "" {
		s.logger.WarnContext(ctx, "get_user: empty user_id")
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		s.logger.WarnContext(ctx, "get_user: invalid user_id", slog.String("user_id", req.UserId))
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	result, err := s.usersService.GetUser(ctx, users.GetUserIn{
		UserID:          userID,
		IncludeSkills:   req.IncludeSkills,
		IncludeContacts: req.IncludeContacts,
	})
	if err != nil {
		return nil, s.handleError(ctx, err, "get_user")
	}

	s.logger.InfoContext(ctx, "get_user: success", slog.String("user_id", req.UserId))
	return mappers.GetUserOutToResponse(result), nil
}
