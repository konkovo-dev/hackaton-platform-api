package hackathonservice

import (
	"context"
	"errors"
	"log/slog"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/hackathon"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HackathonService struct {
	hackathonv1.UnimplementedHackathonServiceServer
	hackathonService *hackathon.Service
	idempotency      *idempotency.Helper
	logger           *slog.Logger
}

var _ hackathonv1.HackathonServiceServer = (*HackathonService)(nil)

func NewHackathonService(
	hackathonService *hackathon.Service,
	idempotencyHelper *idempotency.Helper,
	logger *slog.Logger,
) *HackathonService {
	return &HackathonService{
		hackathonService: hackathonService,
		idempotency:      idempotencyHelper,
		logger:           logger,
	}
}

func (s *HackathonService) handleError(ctx context.Context, err error, operation string) error {
	switch {
	case errors.Is(err, hackathon.ErrHackathonNotFound):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, hackathon.ErrUnauthorized):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, hackathon.ErrForbidden):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, hackathon.ErrEmptyName),
		errors.Is(err, hackathon.ErrEmptyShortDescription),
		errors.Is(err, hackathon.ErrInvalidLocation),
		errors.Is(err, hackathon.ErrMissingStartsAt),
		errors.Is(err, hackathon.ErrMissingEndsAt),
		errors.Is(err, hackathon.ErrMissingRegistrationOpensAt),
		errors.Is(err, hackathon.ErrMissingRegistrationClosesAt),
		errors.Is(err, hackathon.ErrMissingSubmissionsOpensAt),
		errors.Is(err, hackathon.ErrMissingSubmissionsClosesAt),
		errors.Is(err, hackathon.ErrMissingJudgingEndsAt),
		errors.Is(err, hackathon.ErrInvalidDateSequence),
		errors.Is(err, hackathon.ErrInvalidTeamSizeMax),
		errors.Is(err, hackathon.ErrInvalidRegistrationPolicy),
		errors.Is(err, hackathon.ErrInvalidLink):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		s.logger.ErrorContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Internal, "internal error")
	}
}
