package teaminboxservice

import (
	"context"
	"errors"
	"log/slog"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teaminbox"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type API struct {
	teamv1.UnimplementedTeamInboxServiceServer
	teamInboxService *teaminbox.Service
	logger           *slog.Logger
}

var _ teamv1.TeamInboxServiceServer = (*API)(nil)

func New(
	teamInboxService *teaminbox.Service,
	logger *slog.Logger,
) *API {
	return &API{
		teamInboxService: teamInboxService,
		logger:           logger,
	}
}

func (a *API) handleError(ctx context.Context, err error, operation string) error {
	switch {
	case errors.Is(err, teaminbox.ErrUnauthorized):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, teaminbox.ErrForbidden):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, teaminbox.ErrNotFound):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, teaminbox.ErrConflict):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, teaminbox.ErrBadRequest):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		a.logger.ErrorContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Internal, "internal error")
	}
}
