package pingservice

import (
	"context"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
)

type PingService struct {
	hackathonv1.UnimplementedPingServiceServer
}

var _ hackathonv1.PingServiceServer = (*PingService)(nil)

func New() *PingService {
	return &PingService{}
}

func (s *PingService) Ping(ctx context.Context, req *hackathonv1.PingRequest) (*hackathonv1.PingResponse, error) {
	return &hackathonv1.PingResponse{
		Message: "pong",
	}, nil
}
