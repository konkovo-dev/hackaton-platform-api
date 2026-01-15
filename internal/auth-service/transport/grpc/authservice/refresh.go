package authservice

import (
	"context"

	authv1 "github.com/belikoooova/hackaton-platform-api/api/auth/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/transport/grpc/mappers"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/usecase/auth"
)

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
