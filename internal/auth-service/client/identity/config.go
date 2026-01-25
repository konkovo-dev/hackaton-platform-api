package identity

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	IdentityServiceURL string
	ServiceToken       string
}

func NewConfig() (*Config, error) {
	identityServiceURL := env.GetEnv("IDENTITY_SERVICE_URL", "localhost:8081")
	serviceToken := env.GetEnv("SERVICE_AUTH_TOKEN", "")

	if serviceToken == "" {
		return nil, fmt.Errorf("SERVICE_AUTH_TOKEN is required")
	}

	return &Config{
		IdentityServiceURL: identityServiceURL,
		ServiceToken:       serviceToken,
	}, nil
}

func MustNewConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load identity client config: %v", err))
	}
	return cfg
}
