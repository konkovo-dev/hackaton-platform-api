package grpc

import (
	authv1 "github.com/belikoooova/hackaton-platform-api/api/auth/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/transport/grpc/pingservice"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"google.golang.org/grpc"
)

func NewGRPCServer(pingService *pingservice.PingService, authService authv1.AuthServiceServer) *grpc.Server {
	grpcServer := commongrpc.NewServer()

	authv1.RegisterPingServiceServer(grpcServer, pingService)
	authv1.RegisterAuthServiceServer(grpcServer, authService)

	return grpcServer
}
