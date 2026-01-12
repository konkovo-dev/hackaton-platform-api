package pingservice

import (
	"context"

	authv1 "github.com/belikoooova/hackaton-platform-api/api/auth/v1"
)

type PingService struct {
	authv1.UnimplementedPingServiceServer
}

func New() *PingService {
	return &PingService{}
}

func (s *PingService) Ping(ctx context.Context, req *authv1.PingRequest) (*authv1.PingResponse, error) {
	return &authv1.PingResponse{
		Message: "pong",
	}, nil
}
