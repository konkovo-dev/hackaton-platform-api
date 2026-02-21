package participationservice

import (
	"context"
	"errors"
	"log/slog"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/participation"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type API struct {
	participationrolesv1.UnimplementedParticipationServiceServer
	participationService *participation.Service
	teamRoleRepo         participation.TeamRoleRepository
	idempotencyHelper    *idempotency.Helper
	logger               *slog.Logger
}

var _ participationrolesv1.ParticipationServiceServer = (*API)(nil)

func New(
	participationService *participation.Service,
	teamRoleRepo participation.TeamRoleRepository,
	idempotencyHelper *idempotency.Helper,
	logger *slog.Logger,
) *API {
	return &API{
		participationService: participationService,
		teamRoleRepo:         teamRoleRepo,
		idempotencyHelper:    idempotencyHelper,
		logger:               logger,
	}
}

func (a *API) handleError(ctx context.Context, err error, operation string) error {
	switch {
	case errors.Is(err, participation.ErrUnauthorized):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, participation.ErrForbidden):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, participation.ErrInvalidInput):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, participation.ErrConflict):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, participation.ErrNotFound):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.NotFound, err.Error())
	default:
		a.logger.ErrorContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Internal, "internal error")
	}
}
