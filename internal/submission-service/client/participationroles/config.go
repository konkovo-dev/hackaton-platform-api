package participationroles

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	ParticipationRolesServiceURL string
	ServiceToken                 string
}

func NewConfig() *Config {
	return &Config{
		ParticipationRolesServiceURL: env.GetEnv("PARTICIPATION_ROLES_SERVICE_URL", "localhost:50055"),
		ServiceToken:                 env.GetEnv("SERVICE_AUTH_TOKEN", ""),
	}
}

func MustNewConfig() *Config {
	cfg := NewConfig()
	if cfg.ParticipationRolesServiceURL == "" {
		panic(fmt.Errorf("PARTICIPATION_ROLES_SERVICE_URL is required"))
	}
	if cfg.ServiceToken == "" {
		panic(fmt.Errorf("SERVICE_AUTH_TOKEN is required"))
	}
	return cfg
}
