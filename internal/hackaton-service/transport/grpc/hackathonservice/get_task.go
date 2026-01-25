package hackathonservice

import (
	"context"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/hackathon"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *HackathonService) GetHackathonTask(ctx context.Context, req *hackathonv1.GetHackathonTaskRequest) (*hackathonv1.GetHackathonTaskResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid hackathon_id: %v", err)
	}

	in := hackathon.GetTaskIn{
		HackathonID: hackathonID,
	}

	out, err := s.hackathonService.GetTask(ctx, in)
	if err != nil {
		return nil, s.handleError(ctx, err, "GetHackathonTask")
	}

	return &hackathonv1.GetHackathonTaskResponse{
		Task: out.Task,
	}, nil
}
