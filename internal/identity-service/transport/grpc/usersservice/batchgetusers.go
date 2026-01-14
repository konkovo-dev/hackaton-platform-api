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

func (s *UsersService) BatchGetUsers(ctx context.Context, req *identityv1.BatchGetUsersRequest) (*identityv1.BatchGetUsersResponse, error) {
	if len(req.UserIds) == 0 {
		s.logger.WarnContext(ctx, "batch_get_users: empty user_ids")
		return nil, status.Error(codes.InvalidArgument, "user_ids is required")
	}

	userIDs := make([]uuid.UUID, 0, len(req.UserIds))
	for _, idStr := range req.UserIds {
		userID, err := uuid.Parse(idStr)
		if err != nil {
			s.logger.WarnContext(ctx, "batch_get_users: invalid user_id", slog.String("user_id", idStr))
			return nil, status.Error(codes.InvalidArgument, "invalid user_id: "+idStr)
		}
		userIDs = append(userIDs, userID)
	}

	result, err := s.usersService.BatchGetUsers(ctx, users.BatchGetUsersIn{
		UserIDs:         userIDs,
		IncludeSkills:   req.IncludeSkills,
		IncludeContacts: req.IncludeContacts,
	})
	if err != nil {
		return nil, s.handleError(ctx, err, "batch_get_users")
	}

	s.logger.InfoContext(ctx, "batch_get_users: success", slog.Int("count", len(result.Users)))
	return mappers.BatchGetUsersOutToResponse(result), nil
}
