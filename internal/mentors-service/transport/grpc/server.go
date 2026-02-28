package grpc

import (
	"log/slog"

	mentorsv1 "github.com/belikoooova/hackaton-platform-api/api/mentors/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/transport/grpc/mentorsservice"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/interceptor"
	"github.com/belikoooova/hackaton-platform-api/pkg/env"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"google.golang.org/grpc"
)

func NewGRPCServer(
	mentorsAPI *mentorsservice.API,
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

	mentorsv1.RegisterMentorsServiceServer(grpcServer, mentorsAPI)

	return grpcServer
}
