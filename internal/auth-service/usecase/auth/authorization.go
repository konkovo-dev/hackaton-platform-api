package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/domain/entity"
	"github.com/google/uuid"
)

type AuthOut struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

func (s *Service) generateTokens(ctx context.Context, userID uuid.UUID) (*AuthOut, error) {
	accessToken, accessExpiresAt, err := s.jwtService.Sign(userID, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	tokenHash := s.hashRefreshToken(refreshToken)

	refreshExpiresAt := time.Now().UTC().Add(s.cfg.RefreshTokenTTL)
	refreshTokenEntity := &entity.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: refreshExpiresAt,
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &AuthOut{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func (s *Service) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *Service) hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}
