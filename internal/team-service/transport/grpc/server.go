package grpc

import (
	"log/slog"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/transport/grpc/teaminboxservice"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/transport/grpc/teammembersservice"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/transport/grpc/teamsservice"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/interceptor"
	"github.com/belikoooova/hackaton-platform-api/pkg/env"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"google.golang.org/grpc"
)

func NewGRPCServer(
	teamsService *teamsservice.API,
	teamMembersService *teammembersservice.API,
	teamInboxService *teaminboxservice.API,
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

	teamv1.RegisterTeamServiceServer(grpcServer, teamsService)
	teamv1.RegisterTeamMembersServiceServer(grpcServer, teamMembersService)
	teamv1.RegisterTeamInboxServiceServer(grpcServer, teamInboxService)

	return grpcServer
}
