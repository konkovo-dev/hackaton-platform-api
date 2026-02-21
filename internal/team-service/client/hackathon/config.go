package hackathon

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	HackathonServiceURL string
	ServiceToken        string
}

func NewConfig() (*Config, error) {
	serviceURL := env.GetEnv("HACKATHON_SERVICE_URL", "localhost:50051")
	serviceToken := env.GetEnv("SERVICE_AUTH_TOKEN", "")

	if serviceToken == "" {
		return nil, fmt.Errorf("SERVICE_AUTH_TOKEN is required")
	}

	return &Config{
		HackathonServiceURL: serviceURL,
		ServiceToken:        serviceToken,
	}, nil
}

func MustNewConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load hackathon client config: %v", err))
	}
	return cfg
}
