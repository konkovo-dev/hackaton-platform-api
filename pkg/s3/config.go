package s3

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
}

func NewConfig() *Config {
	useSSL := env.GetEnv("S3_USE_SSL", "false") == "true"

	return &Config{
		Endpoint:        env.GetEnv("S3_ENDPOINT", "localhost:9000"),
		Region:          env.GetEnv("S3_REGION", "us-east-1"),
		AccessKeyID:     env.GetEnv("S3_ACCESS_KEY_ID", "minioadmin"),
		SecretAccessKey: env.GetEnv("S3_SECRET_ACCESS_KEY", "minioadmin"),
		BucketName:      env.GetEnv("S3_SUBMISSIONS_BUCKET", "submissions"),
		UseSSL:          useSSL,
	}
}

func MustNewConfig() *Config {
	cfg := NewConfig()

	if cfg.Endpoint == "" {
		panic(fmt.Errorf("S3_ENDPOINT is required"))
	}
	if cfg.AccessKeyID == "" {
		panic(fmt.Errorf("S3_ACCESS_KEY_ID is required"))
	}
	if cfg.SecretAccessKey == "" {
		panic(fmt.Errorf("S3_SECRET_ACCESS_KEY is required"))
	}
	if cfg.BucketName == "" {
		panic(fmt.Errorf("S3_SUBMISSIONS_BUCKET is required"))
	}

	return cfg
}
