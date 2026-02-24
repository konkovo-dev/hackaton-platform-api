package grpc

import (
	mentorsv1 "github.com/belikoooova/hackaton-platform-api/api/mentors/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/transport/grpc/mentorsservice"
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func NewGRPCServer(mentorsAPI *mentorsservice.API) *grpc.Server {
	server := commongrpc.NewServer()

	mentorsv1.RegisterMentorsServiceServer(server, mentorsAPI)

	healthpb.RegisterHealthServer(server, health.NewServer())
	reflection.Register(server)

	return server
}
