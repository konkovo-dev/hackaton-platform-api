package grpc

import (
	"fmt"
	"strconv"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	Port int
}

func NewConfig() (*Config, error) {
	portStr := env.GetEnv("PARTICIPATION_AND_ROLES_SERVICE_GRPC_PORT", "50055")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PARTICIPATION_AND_ROLES_SERVICE_GRPC_PORT: %w", err)
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

