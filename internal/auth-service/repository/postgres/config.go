package postgres

import (
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	DSN               string
	MaxConns          int
	MinConns          int
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

func NewConfig() (*Config, error) {
	dsn := env.GetEnv("DB_DSN", "")
	if dsn == "" {
		return nil, fmt.Errorf("DB_DSN is required")
	}

	maxConns, err := env.GetEnvInt("DB_MAX_CONNS", 25)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_CONNS: %w", err)
	}

	minConns, err := env.GetEnvInt("DB_MIN_CONNS", 5)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MIN_CONNS: %w", err)
	}

	maxConnLifetime, err := env.GetEnvDuration("DB_MAX_CONN_LIFETIME", time.Hour)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_CONN_LIFETIME: %w", err)
	}

	maxConnIdleTime, err := env.GetEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_CONN_IDLE_TIME: %w", err)
	}

	healthCheckPeriod, err := env.GetEnvDuration("DB_HEALTH_CHECK_PERIOD", time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_HEALTH_CHECK_PERIOD: %w", err)
	}

	return &Config{
		DSN:               dsn,
		MaxConns:          maxConns,
		MinConns:          minConns,
		MaxConnLifetime:   maxConnLifetime,
		MaxConnIdleTime:   maxConnIdleTime,
		HealthCheckPeriod: healthCheckPeriod,
	}, nil
}

func MustNewConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load postgres config: %v", err))
	}
	return cfg
}
