package gateway

import (
	"fmt"
	"strconv"

	"github.com/belikoooova/hackaton-platform-api/pkg/env"
)

type Config struct {
	IdentityGRPCEndpoint string
	GatewayHTTPPort      int
}

func NewConfig() *Config {
	identityGRPCEndpoint := env.GetEnv("IDENTITY_GRPC_ENDPOINT", "localhost:50051")
	gatewayHTTPPortStr := env.GetEnv("GATEWAY_HTTP_PORT", "8080")
	gatewayHTTPPort, err := strconv.Atoi(gatewayHTTPPortStr)
	if err != nil {
		panic(fmt.Errorf("failed to convert gateway http port to int: %w", err))
	}

	return &Config{
		IdentityGRPCEndpoint: identityGRPCEndpoint,
		GatewayHTTPPort:      gatewayHTTPPort,
	}
}
