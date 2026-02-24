package participationroles

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	ParticipationRolesServiceURL string
	ServiceToken                 string
}

func NewConfig() (*Config, error) {
	serviceURL := env.GetEnv("PARTICIPATION_ROLES_SERVICE_URL", "localhost:50053")
	serviceToken := env.GetEnv("SERVICE_AUTH_TOKEN", "")

	if serviceToken == "" {
		return nil, fmt.Errorf("SERVICE_AUTH_TOKEN is required")
	}

	return &Config{
		ParticipationRolesServiceURL: serviceURL,
		ServiceToken:                 serviceToken,
	}, nil
}

func MustNewConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load participation-roles client config: %v", err))
	}
	return cfg
}
