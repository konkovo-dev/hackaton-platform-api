package outbox

import (
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	PollingInterval time.Duration
	BatchSize       int
	ProcessTimeout  time.Duration
	MaxAttempts     int
}

func NewConfig() (*Config, error) {
	pollingInterval, err := env.GetEnvDuration("OUTBOX_POLLING_INTERVAL", 1*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_POLLING_INTERVAL: %w", err)
	}

	batchSize, err := env.GetEnvInt("OUTBOX_BATCH_SIZE", 100)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_BATCH_SIZE: %w", err)
	}

	processTimeout, err := env.GetEnvDuration("OUTBOX_PROCESS_TIMEOUT", 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_PROCESS_TIMEOUT: %w", err)
	}

	maxAttempts, err := env.GetEnvInt("OUTBOX_MAX_ATTEMPTS", 3)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_MAX_ATTEMPTS: %w", err)
	}

	return &Config{
		PollingInterval: pollingInterval,
		BatchSize:       batchSize,
		ProcessTimeout:  processTimeout,
		MaxAttempts:     maxAttempts,
	}, nil
}

func MustNewConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load outbox config: %v", err))
	}
	return cfg
}
