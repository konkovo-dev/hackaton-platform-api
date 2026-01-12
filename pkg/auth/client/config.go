package client

import (
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	AuthServiceURL       string
	MaxCacheTTL          time.Duration
	CacheCleanupInterval time.Duration
}

func NewConfig() (*Config, error) {
	authServiceURL := env.GetEnv("AUTH_CLIENT_SERVICE_URL", "localhost:8080")

	maxCacheTTL, err := env.GetEnvDuration("AUTH_CLIENT_MAX_CACHE_TTL", 60*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid AUTH_CLIENT_MAX_CACHE_TTL: %w", err)
	}

	cacheCleanupInterval, err := env.GetEnvDuration("AUTH_CLIENT_CACHE_CLEANUP_INTERVAL", 1*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid AUTH_CLIENT_CACHE_CLEANUP_INTERVAL: %w", err)
	}

	return &Config{
		AuthServiceURL:       authServiceURL,
		MaxCacheTTL:          maxCacheTTL,
		CacheCleanupInterval: cacheCleanupInterval,
	}, nil
}

func MustNewConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load auth client config: %v", err))
	}
	return cfg
}
