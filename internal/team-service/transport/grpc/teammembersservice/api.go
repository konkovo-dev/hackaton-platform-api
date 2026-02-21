package teammembersservice

import (
	"context"
	"errors"
	"log/slog"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teammember"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/vacancy"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type API struct {
	teamv1.UnimplementedTeamMembersServiceServer
	vacancyService    *vacancy.Service
	teamMemberService *teammember.Service
	logger            *slog.Logger
}

var _ teamv1.TeamMembersServiceServer = (*API)(nil)

func New(
	vacancyService *vacancy.Service,
	teamMemberService *teammember.Service,
	logger *slog.Logger,
) *API {
	return &API{
		vacancyService:    vacancyService,
		teamMemberService: teamMemberService,
		logger:            logger,
	}
}

func (a *API) handleError(ctx context.Context, err error, operation string) error {
	switch {
	case errors.Is(err, vacancy.ErrUnauthorized), errors.Is(err, teammember.ErrUnauthorized):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, vacancy.ErrForbidden), errors.Is(err, teammember.ErrForbidden):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, vacancy.ErrNotFound), errors.Is(err, teammember.ErrNotFound):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, vacancy.ErrConflict):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, vacancy.ErrBadRequest):
		a.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		a.logger.ErrorContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Internal, "internal error")
	}
}
