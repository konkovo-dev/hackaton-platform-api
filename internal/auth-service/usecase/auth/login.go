package auth

import (
	"context"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/domain/entity"
)

type LoginIn struct {
	Email    string
	Username string
	Password string
}

func (s *Service) Login(ctx context.Context, in LoginIn) (*AuthOut, error) {
	if err := s.validateLoginIn(in); err != nil {
		return nil, err
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

func (s *Service) validateLoginIn(in LoginIn) error {
	if in.Email == "" && in.Username == "" {
		return ErrEmptyLogin
	}

	if in.Password == "" {
		return ErrEmptyPassword
	}

	return nil
}
