package auth

import (
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func NewConfig() (*Config, error) {
	accessTokenTTL, err := env.GetEnvDuration("ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN_TTL: %w", err)
	}

	refreshTokenTTL, err := env.GetEnvDuration("REFRESH_TOKEN_TTL", 30*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_TTL: %w", err)
	}

	return &Config{
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
	}, nil
}

func MustNewConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load auth config: %v", err))
	}
	return cfg
}
