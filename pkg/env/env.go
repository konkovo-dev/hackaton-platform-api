package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func GetEnvInt(key string, defaultValue int) (int, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}

	return value, nil
}

func GetEnvInt64(key string, defaultValue int64) (int64, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue, nil
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}

	return value, nil
}

func GetEnvDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue, nil
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return time.Duration(0), fmt.Errorf("invalid %s: %w", key, err)
	}

	return value, nil
}
