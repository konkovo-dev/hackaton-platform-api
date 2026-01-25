package auth

import (
	"context"
	"fmt"
	"time"

	authpolicy "github.com/belikoooova/hackaton-platform-api/internal/auth-service/policy"
)

type RefreshIn struct {
	RefreshToken string
}

func (s *Service) Refresh(ctx context.Context, in RefreshIn) (*AuthOut, error) {
	refreshPolicy := authpolicy.NewRefreshPolicy()
	decision := refreshPolicy.ValidateInput(authpolicy.RefreshParams{
		RefreshToken: in.RefreshToken,
	})

	if !decision.Allowed {
		return nil, s.mapPolicyDecisionToError(decision)
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
