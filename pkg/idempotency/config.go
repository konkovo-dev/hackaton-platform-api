package idempotency

import (
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	TTL time.Duration
}

func NewConfig() (*Config, error) {
	ttl, err := env.GetEnvDuration("IDEMPOTENCY_TTL", 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("invalid IDEMPOTENCY_TTL: %w", err)
	}

	return &Config{
		TTL: ttl,
	}, nil
}

func MustNewConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load idempotency config: %v", err))
	}
	return cfg
}
