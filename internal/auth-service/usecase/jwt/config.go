package jwt

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	PrivateKeyPath string
	KeyID          string
	Issuer         string
	Audience       string
}

func NewConfig() (*Config, error) {
	privateKeyPath := env.GetEnv("RS256_PRIVATE_KEY_PATH", "")
	if privateKeyPath == "" {
		return nil, fmt.Errorf("RS256_PRIVATE_KEY_PATH is required")
	}

	return &Config{
		PrivateKeyPath: privateKeyPath,
		KeyID:          env.GetEnv("JWT_KEY_ID", "key-1"),
		Issuer:         env.GetEnv("JWT_ISSUER", "hackaton-platform"),
		Audience:       env.GetEnv("JWT_AUDIENCE", "hackaton-platform"),
	}, nil
}

func MustNewConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load jwt config: %v", err))
	}
	return cfg
}
