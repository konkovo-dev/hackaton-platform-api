package grpc

import (
	"log/slog"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/transport/grpc/hackathonservice"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/transport/grpc/pingservice"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/interceptor"
	"github.com/belikoooova/hackaton-platform-api/pkg/env"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"google.golang.org/grpc"
)

func NewGRPCServer(
	pingService *pingservice.PingService,
	hackathonService *hackathonservice.HackathonService,
	authClient client.AuthClient,
	logger *slog.Logger,
) *grpc.Server {
	publicMethods := []string{
		"/hackathon.v1.PingService/Ping",
	}

	optionalMethods := []string{
		"/hackathon.v1.HackathonService/ListHackathons",
	}

	internalMethods := []string{}

	hybridMethods := []string{
		"/hackathon.v1.HackathonService/GetHackathon",
	}

	serviceToken := env.GetEnv("SERVICE_AUTH_TOKEN", "")

	authInterceptor := interceptor.NewUnaryInterceptorWithHybrid(
		authClient,
		publicMethods,
		optionalMethods,
		internalMethods,
		hybridMethods,
		serviceToken,
		logger,
	)

	grpcServer := commongrpc.NewServer(commongrpc.ServerOptions{
		UnaryInterceptors: []grpc.UnaryServerInterceptor{authInterceptor},
	})

	hackathonv1.RegisterPingServiceServer(grpcServer, pingService)
	hackathonv1.RegisterHackathonServiceServer(grpcServer, hackathonService)

	return grpcServer
}
