package staffservice

import (
	"context"
	"errors"
	"log/slog"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/role"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type API struct {
	participationrolesv1.UnimplementedStaffServiceServer
	roleService       *role.Service
	idempotencyHelper *idempotency.Helper
	logger            *slog.Logger
}

var _ participationrolesv1.StaffServiceServer = (*API)(nil)

func New(
	roleService *role.Service,
	idempotencyHelper *idempotency.Helper,
	logger *slog.Logger,
) *API {
	return &API{
		roleService:       roleService,
		idempotencyHelper: idempotencyHelper,
		logger:            logger,
	}
}

func (a *API) handleError(ctx context.Context, err error, operation string) error {
	switch {
	case errors.Is(err, role.ErrUnauthorized):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, role.ErrForbidden):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, role.ErrInvalidInput):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, role.ErrHackathonNotFound):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, role.ErrUserNotFound):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, role.ErrConflict):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, role.ErrNotFound):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.NotFound, err.Error())
	default:
		a.logger.ErrorContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Internal, "internal error")
	}
}
