package pingservice

import (
	"context"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
)

type PingService struct {
	identityv1.UnimplementedPingServiceServer
}

func New() *PingService {
	return &PingService{}
}

func (s *PingService) Ping(ctx context.Context, req *identityv1.PingRequest) (*identityv1.PingResponse, error) {
	return &identityv1.PingResponse{
		Message: "pong",
	}, nil
}
