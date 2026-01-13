package grpc

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	Port int
}

type ConfigOptions struct {
	EnvVarName  string
	DefaultPort int
}

func NewConfig(opts ConfigOptions) (*Config, error) {
	port, err := env.GetEnvInt(opts.EnvVarName, opts.DefaultPort)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", opts.EnvVarName, err)
	}

	return &Config{Port: port}, nil
}

func MustNewConfig(opts ConfigOptions) *Config {
	cfg, err := NewConfig(opts)
	if err != nil {
		panic(fmt.Sprintf("failed to load grpc config: %v", err))
	}
	return cfg
}
