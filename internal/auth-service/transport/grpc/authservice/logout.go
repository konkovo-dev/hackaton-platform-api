package authservice

import (
	"context"

	authv1 "github.com/belikoooova/hackaton-platform-api/api/auth/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/usecase/auth"
)

func (s *AuthService) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := s.authService.Logout(ctx, auth.LogoutIn{
		RefreshToken: req.RefreshToken,
	}); err != nil {
		return nil, s.handleAuthError(ctx, err, "logout")
	}

	s.logger.InfoContext(ctx, "user logged out")

	return &authv1.LogoutResponse{}, nil
}
