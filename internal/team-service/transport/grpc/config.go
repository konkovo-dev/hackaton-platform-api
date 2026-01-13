package grpc

import (
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
)

func NewConfig() (*commongrpc.Config, error) {
	return commongrpc.NewConfig(commongrpc.ConfigOptions{
		EnvVarName:  "TEAM_SERVICE_GRPC_PORT",
		DefaultPort: 50053,
	})
}

func MustNewConfig() *commongrpc.Config {
	return commongrpc.MustNewConfig(commongrpc.ConfigOptions{
		EnvVarName:  "TEAM_SERVICE_GRPC_PORT",
		DefaultPort: 50053,
	})
}
