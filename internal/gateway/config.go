package gateway

import (
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	IdentityGRPCEndpoint string
	GatewayHTTPPort      int
}

func NewConfig() *Config {
	identityGRPCEndpoint := env.GetEnv("IDENTITY_GRPC_ENDPOINT", "localhost:50051")

	gatewayHTTPPort, err := env.GetEnvInt("GATEWAY_HTTP_PORT", 8080)
	if err != nil {
		panic(fmt.Errorf("invalid GATEWAY_HTTP_PORT: %w", err))
	}

	return &Config{
		IdentityGRPCEndpoint: identityGRPCEndpoint,
		GatewayHTTPPort:      gatewayHTTPPort,
	}
}
