package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
)

type AuthClient interface {
	IntrospectToken(ctx context.Context, token string) (*auth.Claims, error)
}

func NewAuthClient(cfg *Config) (AuthClient, error) {
	authClient, err := newAuthClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth client: %w", err)
	}

	return newCachingAuthClient(cfg, authClient), nil
}

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrInvalidToken  = errors.New("invalid token")
)
