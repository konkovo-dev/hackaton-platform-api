package hackathon

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	HackathonServiceURL string
	ServiceToken        string
}

func MustNewConfig() *Config {
	hackathonServiceURL := env.GetEnv("HACKATHON_CLIENT_SERVICE_URL", "localhost:50052")
	serviceToken := env.GetEnv("SERVICE_AUTH_TOKEN", "")

	if serviceToken == "" {
		panic(fmt.Errorf("SERVICE_AUTH_TOKEN is required"))
	}

	return &Config{
		HackathonServiceURL: hackathonServiceURL,
		ServiceToken:        serviceToken,
	}
}
