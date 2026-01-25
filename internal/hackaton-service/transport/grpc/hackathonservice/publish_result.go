package hackathonservice

import (
	"context"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/hackathon"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *HackathonService) PublishHackathonResult(ctx context.Context, req *hackathonv1.PublishHackathonResultRequest) (*hackathonv1.PublishHackathonResultResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid hackathon_id: %v", err)
	}

	in := hackathon.PublishResultIn{
		HackathonID: hackathonID,
	}

	_, err = s.hackathonService.PublishResult(ctx, in)
	if err != nil {
		return nil, s.handleError(ctx, err, "PublishHackathonResult")
	}

	return &hackathonv1.PublishHackathonResultResponse{}, nil
}
