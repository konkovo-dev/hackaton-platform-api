package auth

import (
	"context"
	"fmt"
	"time"

	authpolicy "github.com/belikoooova/hackaton-platform-api/internal/auth-service/policy"
)

type LogoutIn struct {
	RefreshToken string
}

func (s *Service) Logout(ctx context.Context, in LogoutIn) error {
	logoutPolicy := authpolicy.NewLogoutPolicy()
	decision := logoutPolicy.ValidateInput(authpolicy.LogoutParams{
		RefreshToken: in.RefreshToken,
	})

	if !decision.Allowed {
		return s.mapPolicyDecisionToError(decision)
	}

	tokenHash := s.hashRefreshToken(in.RefreshToken)

	if err := s.refreshTokenRepo.Revoke(ctx, tokenHash, time.Now().UTC()); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}
