package authservice

import (
	"context"

	authv1 "github.com/belikoooova/hackaton-platform-api/api/auth/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/transport/grpc/mappers"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/usecase/auth"
)

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
