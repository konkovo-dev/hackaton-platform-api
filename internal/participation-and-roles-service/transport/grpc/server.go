package grpc

import (
	"log/slog"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/transport/grpc/participationandrolesservice"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/interceptor"
	"github.com/belikoooova/hackaton-platform-api/pkg/env"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"google.golang.org/grpc"
)

func NewGRPCServer(
	parService *participationandrolesservice.API,
	authClient client.AuthClient,
	logger *slog.Logger,
) *grpc.Server {
	publicMethods := []string{}

	optionalMethods := []string{}

	internalMethods := []string{
		"/participationandroles.v1.ParticipationAndRolesService/AssignHackathonRole",
	}

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

	participationrolesv1.RegisterParticipationAndRolesServiceServer(grpcServer, parService)

	return grpcServer
}
