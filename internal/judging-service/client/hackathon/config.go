package hackathon

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	HackathonServiceURL string
	ServiceToken        string
}

func NewConfig() *Config {
	return &Config{
		HackathonServiceURL: env.GetEnv("HACKATHON_SERVICE_URL", "localhost:50052"),
		ServiceToken:        env.GetEnv("SERVICE_AUTH_TOKEN", ""),
	}
}

func MustNewConfig() *Config {
	cfg := NewConfig()
	if cfg.HackathonServiceURL == "" {
		panic(fmt.Errorf("HACKATHON_SERVICE_URL is required"))
	}
	if cfg.ServiceToken == "" {
		panic(fmt.Errorf("SERVICE_AUTH_TOKEN is required"))
	}
	return cfg
}
