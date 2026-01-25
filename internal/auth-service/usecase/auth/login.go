package auth

import (
	"context"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/domain/entity"
	authpolicy "github.com/belikoooova/hackaton-platform-api/internal/auth-service/policy"
)

type LoginIn struct {
	Email    string
	Username string
	Password string
}

func (s *Service) Login(ctx context.Context, in LoginIn) (*AuthOut, error) {
	loginPolicy := authpolicy.NewLoginPolicy()
	decision := loginPolicy.ValidateInput(authpolicy.LoginParams{
		Email:    in.Email,
		Username: in.Username,
		Password: in.Password,
	})

	if !decision.Allowed {
		return nil, s.mapPolicyDecisionToError(decision)
	}

	var user *entity.User
	var err error

	if in.Email != "" {
		user, err = s.userRepo.GetByEmail(ctx, in.Email)
	} else {
		user, err = s.userRepo.GetByUsername(ctx, strings.ToLower(in.Username))
	}

	if err != nil {
		return nil, ErrInvalidCredentials
	}

	credentials, err := s.credentialsRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := s.passwordService.Verify(in.Password, credentials.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokens(ctx, user.ID)
}
