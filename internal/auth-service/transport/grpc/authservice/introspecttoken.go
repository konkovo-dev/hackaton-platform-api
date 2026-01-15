package authservice

import (
	"context"
	"errors"
	"log/slog"

	authv1 "github.com/belikoooova/hackaton-platform-api/api/auth/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/transport/grpc/mappers"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/usecase/auth"
)

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

