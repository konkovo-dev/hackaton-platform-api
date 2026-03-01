package team

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	TeamServiceURL string
	ServiceToken   string
}

func NewConfig() *Config {
	return &Config{
		TeamServiceURL: env.GetEnv("TEAM_SERVICE_URL", "localhost:50053"),
		ServiceToken:   env.GetEnv("SERVICE_AUTH_TOKEN", ""),
	}
}

func MustNewConfig() *Config {
	cfg := NewConfig()
	if cfg.TeamServiceURL == "" {
		panic(fmt.Errorf("TEAM_SERVICE_URL is required"))
	}
	if cfg.ServiceToken == "" {
		panic(fmt.Errorf("SERVICE_AUTH_TOKEN is required"))
	}
	return cfg
}
