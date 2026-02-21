package grpc

import (
	"log/slog"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/transport/grpc/participationservice"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/transport/grpc/staffservice"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/interceptor"
	"github.com/belikoooova/hackaton-platform-api/pkg/env"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"google.golang.org/grpc"
)

func NewGRPCServer(
	staffSvc *staffservice.API,
	participationSvc *participationservice.API,
	authClient client.AuthClient,
	logger *slog.Logger,
) *grpc.Server {
	publicMethods := []string{}

	optionalMethods := []string{}

	internalMethods := []string{
		"/participationandroles.v1.StaffService/AssignHackathonRole",
		"/participationandroles.v1.ParticipationService/ConvertToTeamParticipation",
		"/participationandroles.v1.ParticipationService/ConvertFromTeamParticipation",
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

	participationrolesv1.RegisterStaffServiceServer(grpcServer, staffSvc)
	participationrolesv1.RegisterParticipationServiceServer(grpcServer, participationSvc)

	return grpcServer
}
