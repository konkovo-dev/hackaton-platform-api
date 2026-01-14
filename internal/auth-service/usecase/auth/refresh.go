package auth

import (
	"context"
	"fmt"
	"time"
)

type RefreshIn struct {
	RefreshToken string
}

func (s *Service) Refresh(ctx context.Context, in RefreshIn) (*AuthOut, error) {
	if err := s.validateRefreshIn(in); err != nil {
		return nil, err
	}

	tokenHash := s.hashRefreshToken(in.RefreshToken)

	storedToken, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	if storedToken.RevokedAt != nil {
		return nil, ErrTokenRevoked
	}
	if time.Now().UTC().After(storedToken.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	user, err := s.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if err := s.refreshTokenRepo.Revoke(ctx, tokenHash, time.Now().UTC()); err != nil {
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	return s.generateTokens(ctx, user.ID)
}

func (s *Service) validateRefreshIn(in RefreshIn) error {
	if in.RefreshToken == "" {
		return ErrTokenInvalid
	}

	return nil
}
