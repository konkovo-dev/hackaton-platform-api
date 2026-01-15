package grpc

import (
	"log/slog"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/meservice"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/pingservice"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/skillsservice"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/usersservice"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth/interceptor"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"google.golang.org/grpc"
)

func NewGRPCServer(
	pingService *pingservice.PingService,
	meService *meservice.MeService,
	usersService *usersservice.UsersService,
	skillsService *skillsservice.SkillsService,
	authClient client.AuthClient,
	logger *slog.Logger,
) *grpc.Server {
	publicMethods := []string{
		"/identity.v1.PingService/Ping",
		"/identity.v1.MeService/CreateMe",
	}

	authInterceptor := interceptor.NewUnaryInterceptor(authClient, publicMethods, logger)

	grpcServer := commongrpc.NewServer(commongrpc.ServerOptions{
		UnaryInterceptors: []grpc.UnaryServerInterceptor{authInterceptor},
	})

	identityv1.RegisterPingServiceServer(grpcServer, pingService)
	identityv1.RegisterMeServiceServer(grpcServer, meService)
	identityv1.RegisterUsersServiceServer(grpcServer, usersService)
	identityv1.RegisterSkillsServiceServer(grpcServer, skillsService)

	return grpcServer
}
