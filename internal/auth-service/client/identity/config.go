package identity

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	IdentityServiceURL string
}

func NewConfig() (*Config, error) {
	identityServiceURL := env.GetEnv("IDENTITY_SERVICE_URL", "localhost:8081")

	return &Config{
		IdentityServiceURL: identityServiceURL,
	}, nil
}

func MustNewConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load identity client config: %v", err))
	}
	return cfg
}
