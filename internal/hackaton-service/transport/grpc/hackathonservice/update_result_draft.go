package hackathonservice

import (
	"context"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/hackathon"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *HackathonService) UpdateHackathonResultDraft(ctx context.Context, req *hackathonv1.UpdateHackathonResultDraftRequest) (*hackathonv1.UpdateHackathonResultDraftResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid hackathon_id: %v", err)
	}

	in := hackathon.UpdateResultDraftIn{
		HackathonID: hackathonID,
		Result:      req.Result,
	}

	_, err = s.hackathonService.UpdateResultDraft(ctx, in)
	if err != nil {
		return nil, s.handleError(ctx, err, "UpdateHackathonResultDraft")
	}

	return &hackathonv1.UpdateHackathonResultDraftResponse{}, nil
}

