package centrifugo

import (
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	URL       string
	APIKey    string
	JWTSecret string
	JWTTTL    time.Duration
}

func NewConfig() (*Config, error) {
	url := env.GetEnv("CENTRIFUGO_URL", "http://localhost:8000")
	apiKey := env.GetEnv("CENTRIFUGO_API_KEY", "dev-api-key-change-in-production")
	jwtSecret := env.GetEnv("CENTRIFUGO_JWT_SECRET", "dev-jwt-secret-change-in-production")
	jwtTTLStr := env.GetEnv("CENTRIFUGO_JWT_TTL", "24h")

	jwtTTL, err := time.ParseDuration(jwtTTLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid CENTRIFUGO_JWT_TTL: %w", err)
	}

	return &Config{
		URL:       url,
		APIKey:    apiKey,
		JWTSecret: jwtSecret,
		JWTTTL:    jwtTTL,
	}, nil
}

func MustNewConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load Centrifugo config: %v", err))
	}
	return cfg
}
