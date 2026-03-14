package submission

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	SubmissionServiceURL string
	ServiceToken         string
}

func NewConfig() *Config {
	return &Config{
		SubmissionServiceURL: env.GetEnv("SUBMISSION_SERVICE_URL", "localhost:50058"),
		ServiceToken:         env.GetEnv("SERVICE_AUTH_TOKEN", ""),
	}
}

func MustNewConfig() *Config {
	cfg := NewConfig()
	if cfg.SubmissionServiceURL == "" {
		panic(fmt.Errorf("SUBMISSION_SERVICE_URL is required"))
	}
	if cfg.ServiceToken == "" {
		panic(fmt.Errorf("SERVICE_AUTH_TOKEN is required"))
	}
	return cfg
}
