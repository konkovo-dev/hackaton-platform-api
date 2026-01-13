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
	"google.golang.org/protobuf/types/known/timestamppb"
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

	if req.Username == "" {
		s.logger.WarnContext(ctx, "register: empty username")
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	if req.Password == "" {
		s.logger.WarnContext(ctx, "register: empty password")
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetEmail() == "" {
		s.logger.WarnContext(ctx, "register: empty email")
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.FirstName == "" {
		s.logger.WarnContext(ctx, "register: empty first_name")
		return nil, status.Error(codes.InvalidArgument, "first_name is required")
	}

	if req.LastName == "" {
		s.logger.WarnContext(ctx, "register: empty last_name")
		return nil, status.Error(codes.InvalidArgument, "last_name is required")
	}

	if req.Timezone == "" {
		s.logger.WarnContext(ctx, "register: empty timezone")
		return nil, status.Error(codes.InvalidArgument, "timezone is required")
	}

	result, err := s.authService.Register(ctx, auth.RegisterInput{
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

	resp := &authv1.RegisterResponse{
		AccessToken:      result.AccessToken,
		RefreshToken:     result.RefreshToken,
		AccessExpiresAt:  timestamppb.New(result.AccessExpiresAt),
		RefreshExpiresAt: timestamppb.New(result.RefreshExpiresAt),
	}

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
	var login string
	if email := req.GetEmail(); email != "" {
		login = email
	} else if username := req.GetUsername(); username != "" {
		login = username
	} else {
		s.logger.WarnContext(ctx, "login: empty login")
		return nil, status.Error(codes.InvalidArgument, "email or username is required")
	}

	if req.Password == "" {
		s.logger.WarnContext(ctx, "login: empty password")
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	result, err := s.authService.Login(ctx, login, req.Password)
	if err != nil {
		return nil, s.handleAuthError(ctx, err, "login")
	}

	s.logger.InfoContext(ctx, "user logged in", slog.String("login", login))

	return &authv1.LoginResponse{
		AccessToken:      result.AccessToken,
		RefreshToken:     result.RefreshToken,
		AccessExpiresAt:  timestamppb.New(result.AccessExpiresAt),
		RefreshExpiresAt: timestamppb.New(result.RefreshExpiresAt),
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	if req.RefreshToken == "" {
		s.logger.WarnContext(ctx, "refresh: empty refresh_token")
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	result, err := s.authService.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, s.handleAuthError(ctx, err, "refresh")
	}

	s.logger.InfoContext(ctx, "token refreshed")

	return &authv1.RefreshResponse{
		AccessToken:      result.AccessToken,
		RefreshToken:     result.RefreshToken,
		AccessExpiresAt:  timestamppb.New(result.AccessExpiresAt),
		RefreshExpiresAt: timestamppb.New(result.RefreshExpiresAt),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if req.RefreshToken == "" {
		s.logger.WarnContext(ctx, "logout: empty refresh_token")
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	if err := s.authService.Logout(ctx, req.RefreshToken); err != nil {
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

	active, userID, expiresAt, err := s.authService.IntrospectToken(ctx, req.AccessToken)
	if err != nil {
		if !errors.Is(err, auth.ErrTokenInvalid) && !errors.Is(err, auth.ErrTokenExpired) {
			s.logger.WarnContext(ctx, "introspect token error", slog.String("error", err.Error()))
		}
		return &authv1.IntrospectTokenResponse{
			Active: false,
		}, nil
	}

	return &authv1.IntrospectTokenResponse{
		Active:    active,
		UserId:    userID.String(),
		ExpiresAt: timestamppb.New(expiresAt),
	}, nil
}

func (s *AuthService) handleAuthError(ctx context.Context, err error, operation string) error {
	switch {
	case errors.Is(err, auth.ErrUserAlreadyExists):
		s.logger.WarnContext(ctx, operation+": user already exists")
		return status.Error(codes.AlreadyExists, "user already exists")
	case errors.Is(err, auth.ErrInvalidCredentials):
		s.logger.WarnContext(ctx, operation+": invalid credentials")
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, auth.ErrTokenInvalid):
		s.logger.WarnContext(ctx, operation+": invalid token")
		return status.Error(codes.Unauthenticated, "invalid token")
	case errors.Is(err, auth.ErrTokenExpired):
		s.logger.WarnContext(ctx, operation+": token expired")
		return status.Error(codes.Unauthenticated, "token expired")
	case errors.Is(err, auth.ErrTokenRevoked):
		s.logger.WarnContext(ctx, operation+": token revoked")
		return status.Error(codes.Unauthenticated, "token revoked")
	case errors.Is(err, auth.ErrInvalidUsername):
		s.logger.WarnContext(ctx, operation+": invalid username")
		return status.Error(codes.InvalidArgument, "invalid username")
	case errors.Is(err, auth.ErrInvalidPassword):
		s.logger.WarnContext(ctx, operation+": invalid password")
		return status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	default:
		s.logger.ErrorContext(ctx, operation+": internal error", slog.String("error", err.Error()))
		return status.Error(codes.Internal, "internal error")
	}
}
