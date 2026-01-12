package client

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
)

type cachingAuthClient struct {
	cache      cache
	authClient AuthClient
	maxTTL     time.Duration
	logger     *slog.Logger
}

func newCachingAuthClient(cfg *Config, authClient AuthClient) *cachingAuthClient {
	return &cachingAuthClient{
		cache:      newCache(cfg),
		authClient: authClient,
		maxTTL:     cfg.MaxCacheTTL,
		logger:     slog.Default(),
	}
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (c *cachingAuthClient) IntrospectToken(ctx context.Context, token string) (*auth.Claims, error) {
	tokenHash := hashToken(token)

	claims, err := c.cache.Get(tokenHash)
	if err != nil {
		if !errors.Is(err, ErrTokenNotFound) {
			c.logger.Error("failed to get token from cache", "error", err)
		}

		claims, err = c.authClient.IntrospectToken(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("failed to introspect token: %w", err)
		}

		ttl := time.Until(claims.ExpiresAt)
		if ttl > c.maxTTL {
			ttl = c.maxTTL
		}

		if ttl > 0 {
			if err := c.cache.Set(tokenHash, claims, ttl); err != nil {
				c.logger.Error("failed to set token in cache", "error", err)
			}
		}
	}

	return claims, nil
}
