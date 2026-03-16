package meservice

import (
	"context"
	"errors"
	"log/slog"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/me"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MeService struct {
	identityv1.UnimplementedMeServiceServer
	meService   *me.Service
	idempotency *idempotency.Helper
	logger      *slog.Logger
}

var _ identityv1.MeServiceServer = (*MeService)(nil)

func NewMeService(
	meService *me.Service,
	idempotencyHelper *idempotency.Helper,
	logger *slog.Logger,
) *MeService {
	return &MeService{
		meService:   meService,
		idempotency: idempotencyHelper,
		logger:      logger,
	}
}

func (s *MeService) handleError(ctx context.Context, err error, operation string) error {
	switch {
	case errors.Is(err, me.ErrUserNotFound):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, me.ErrUserAlreadyExists):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, me.ErrInvalidInput):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, me.ErrUnauthorized):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, me.ErrForbidden):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		s.logger.ErrorContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Internal, "internal error")
	}
}
