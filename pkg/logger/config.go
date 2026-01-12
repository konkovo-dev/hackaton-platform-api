package logger

import (
	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	Env string
}

func NewConfig() *Config {
	env := env.GetEnv("ENV", "local")
	return &Config{
		Env: env,
	}
}
