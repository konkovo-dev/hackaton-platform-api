package grpc

import (
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
)

// NewConfig creates a new gRPC configuration for submission-service
func NewConfig() (*commongrpc.Config, error) {
	return commongrpc.NewConfig(commongrpc.ConfigOptions{
		EnvVarName:  "SUBMISSION_SERVICE_GRPC_PORT",
		DefaultPort: 50054,
	})
}

// MustNewConfig creates a new gRPC configuration and panics on error
func MustNewConfig() *commongrpc.Config {
	return commongrpc.MustNewConfig(commongrpc.ConfigOptions{
		EnvVarName:  "SUBMISSION_SERVICE_GRPC_PORT",
		DefaultPort: 50054,
	})
}
