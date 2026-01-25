package identity

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	IdentityServiceURL string
	ServiceToken       string
}

func MustNewConfig() *Config {
	identityServiceURL := env.GetEnv("IDENTITY_CLIENT_SERVICE_URL", "localhost:50051")
	serviceToken := env.GetEnv("SERVICE_AUTH_TOKEN", "")

	if serviceToken == "" {
		panic(fmt.Errorf("SERVICE_AUTH_TOKEN is required"))
	}

	return &Config{
		IdentityServiceURL: identityServiceURL,
		ServiceToken:       serviceToken,
	}
}
