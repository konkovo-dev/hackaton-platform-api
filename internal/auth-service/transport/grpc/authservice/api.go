package authservice

import (
	"context"
	"errors"
	"log/slog"

	authv1 "github.com/belikoooova/hackaton-platform-api/api/auth/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/transport/grpc/mappers"
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

func (s *AuthService) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.Username, req.GetEmail())
		resp := &authv1.RegisterResponse{}
		found, err := s.idempotency.CheckAndGet(ctx, idempotencyKey, "register", requestHash, resp)
		if err != nil {
			var conflictErr *idempotency.ConflictError
			if errors.As(err, &conflictErr) {
				s.logger.WarnContext(ctx, "idempotency key conflict", slog.String("key", idempotencyKey))
				return nil, status.Error(codes.AlreadyExists, "idempotency key already used with different request")
			}
			s.logger.ErrorContext(ctx, "failed to check idempotency", slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, "failed to check idempotency")
		}
		if found {
			s.logger.InfoContext(ctx, "returning cached response", slog.String("key", idempotencyKey))
			return resp, nil
		}
	}

	result, err := s.authService.Register(ctx, auth.RegisterIn{
		Email:     req.GetEmail(),
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Timezone:  req.Timezone,
	})

	if err != nil {
		return nil, s.handleAuthError(ctx, err, "register")
	}

	resp := mappers.AuthOutToRegisterResponse(result)

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.Username, req.GetEmail())
		if err := s.idempotency.Save(ctx, idempotencyKey, "register", requestHash, resp); err != nil {
			s.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	s.logger.InfoContext(ctx, "user registered", slog.String("username", req.Username))
	return resp, nil
}

func (s *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	result, err := s.authService.Login(ctx, auth.LoginIn{
		Email:    req.GetEmail(),
		Username: req.GetUsername(),
		Password: req.Password,
	})
	if err != nil {
		return nil, s.handleAuthError(ctx, err, "login")
	}

	s.logger.InfoContext(ctx, "user logged in")

	return mappers.AuthOutToLoginResponse(result), nil
}

func (s *AuthService) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	result, err := s.authService.Refresh(ctx, auth.RefreshIn{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, s.handleAuthError(ctx, err, "refresh")
	}

	s.logger.InfoContext(ctx, "token refreshed")

	return mappers.AuthOutToRefreshResponse(result), nil
}

func (s *AuthService) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := s.authService.Logout(ctx, auth.LogoutIn{
		RefreshToken: req.RefreshToken,
	}); err != nil {
		return nil, s.handleAuthError(ctx, err, "logout")
	}

	s.logger.InfoContext(ctx, "user logged out")

	return &authv1.LogoutResponse{}, nil
}

func (s *AuthService) IntrospectToken(ctx context.Context, req *authv1.IntrospectTokenRequest) (*authv1.IntrospectTokenResponse, error) {
	if req.AccessToken == "" {
		return &authv1.IntrospectTokenResponse{
			Active: false,
		}, nil
	}

	result, err := s.authService.IntrospectToken(ctx, auth.IntrospectTokenIn{
		AccessToken: req.AccessToken,
	})
	if err != nil {
		if !errors.Is(err, auth.ErrTokenInvalid) && !errors.Is(err, auth.ErrTokenExpired) {
			s.logger.WarnContext(ctx, "introspect token error", slog.String("error", err.Error()))
		}
		return &authv1.IntrospectTokenResponse{
			Active: false,
		}, nil
	}

	return mappers.IntrospectTokenOutToResponse(result), nil
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
