package authservice

import (
	"context"
	"errors"
	"log/slog"

	authv1 "github.com/belikoooova/hackaton-platform-api/api/auth/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/usecase/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService struct {
	authv1.UnimplementedAuthServiceServer
	authService *auth.Service
	idempotency *idempotency.Helper
	logger      *slog.Logger
}

func NewAuthService(
	authService *auth.Service,
	idempotencyHelper *idempotency.Helper,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		authService: authService,
		idempotency: idempotencyHelper,
		logger:      logger,
	}
}

func (s *AuthService) handleAuthError(ctx context.Context, err error, operation string) error {
	switch {
	case errors.Is(err, auth.ErrUserAlreadyExists):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, auth.ErrInvalidCredentials):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, auth.ErrTokenInvalid),
		errors.Is(err, auth.ErrTokenExpired),
		errors.Is(err, auth.ErrTokenRevoked):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, auth.ErrInvalidUsername),
		errors.Is(err, auth.ErrInvalidPassword),
		errors.Is(err, auth.ErrInvalidEmail),
		errors.Is(err, auth.ErrEmptyUsername),
		errors.Is(err, auth.ErrEmptyEmail),
		errors.Is(err, auth.ErrEmptyPassword),
		errors.Is(err, auth.ErrEmptyFirstName),
		errors.Is(err, auth.ErrEmptyLastName),
		errors.Is(err, auth.ErrEmptyLogin),
		errors.Is(err, auth.ErrEmptyTimezone):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		s.logger.ErrorContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Internal, "internal error")
	}
}
