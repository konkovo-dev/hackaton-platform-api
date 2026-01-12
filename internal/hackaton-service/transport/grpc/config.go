package grpc

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	Port int
}

func NewConfig() (*Config, error) {
	port, err := env.GetEnvInt("HACKATON_SERVICE_GRPC_PORT", 50052)
	if err != nil {
		return nil, fmt.Errorf("invalid HACKATON_SERVICE_GRPC_PORT: %w", err)
	}

	return &Config{
		Port: port,
	}, nil
}

func MustNewConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load grpc config: %v", err))
	}
	return cfg
}
