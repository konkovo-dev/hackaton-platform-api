package usersservice

import (
	"context"
	"errors"
	"log/slog"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/users"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UsersService struct {
	identityv1.UnimplementedUsersServiceServer
	usersService *users.Service
	logger       *slog.Logger
}

var _ identityv1.UsersServiceServer = (*UsersService)(nil)

func NewUsersService(
	usersService *users.Service,
	logger *slog.Logger,
) *UsersService {
	return &UsersService{
		usersService: usersService,
		logger:       logger,
	}
}

func (s *UsersService) handleError(ctx context.Context, err error, operation string) error {
	switch {
	case errors.Is(err, users.ErrUserNotFound):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, users.ErrInvalidInput):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, users.ErrTooManyUsers):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		s.logger.ErrorContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Internal, "internal error")
	}
}
