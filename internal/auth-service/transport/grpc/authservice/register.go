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
