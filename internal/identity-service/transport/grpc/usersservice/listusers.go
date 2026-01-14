package usersservice

import (
	"context"
	"log/slog"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/mappers"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/users"
)

func (s *UsersService) ListUsers(ctx context.Context, req *identityv1.ListUsersRequest) (*identityv1.ListUsersResponse, error) {
	result, err := s.usersService.ListUsers(ctx, users.ListUsersIn{
		Query:           req.Query,
		IncludeSkills:   req.IncludeSkills,
		IncludeContacts: req.IncludeContacts,
	})
	if err != nil {
		return nil, s.handleError(ctx, err, "list_users")
	}

	s.logger.InfoContext(ctx, "list_users: success", slog.Int("count", len(result.Users)))
	return mappers.ListUsersOutToResponse(result), nil
}
