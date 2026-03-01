package grpc

import (
	"log/slog"

	matchmakingv1 "github.com/belikoooova/hackaton-platform-api/api/matchmaking/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/transport/grpc/matchmakingservice"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/interceptor"
	"github.com/belikoooova/hackaton-platform-api/pkg/env"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"google.golang.org/grpc"
)

func NewGRPCServer(
	matchmakingAPI *matchmakingservice.API,
	authClient client.AuthClient,
	logger *slog.Logger,
) *grpc.Server {
	publicMethods := []string{}

	optionalMethods := []string{}

	internalMethods := []string{}

	serviceToken := env.GetEnv("SERVICE_AUTH_TOKEN", "")

	authInterceptor := interceptor.NewUnaryInterceptor(
		authClient,
		publicMethods,
		optionalMethods,
		internalMethods,
		serviceToken,
		logger,
	)

	grpcServer := commongrpc.NewServer(commongrpc.ServerOptions{
		UnaryInterceptors: []grpc.UnaryServerInterceptor{authInterceptor},
	})

	matchmakingv1.RegisterMatchmakingServiceServer(grpcServer, matchmakingAPI)

	return grpcServer
}
