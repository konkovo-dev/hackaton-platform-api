package auth

import (
	"context"
	"fmt"
	"time"
)

type LogoutIn struct {
	RefreshToken string
}

func (s *Service) Logout(ctx context.Context, in LogoutIn) error {
	if err := s.validateLogoutIn(in); err != nil {
		return err
	}

	tokenHash := s.hashRefreshToken(in.RefreshToken)

	if err := s.refreshTokenRepo.Revoke(ctx, tokenHash, time.Now().UTC()); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

func (s *Service) validateLogoutIn(in LogoutIn) error {
	if in.RefreshToken == "" {
		return ErrTokenInvalid
	}

	return nil
}
