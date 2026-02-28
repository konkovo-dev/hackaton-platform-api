package nats

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	URL string
}

func NewConfig() (*Config, error) {
	url := env.GetEnv("NATS_URL", "nats://localhost:4222")

	return &Config{
		URL: url,
	}, nil
}

func MustNewConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load NATS config: %v", err))
	}
	return cfg
}
