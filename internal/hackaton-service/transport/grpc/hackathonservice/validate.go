package hackathonservice

import (
	"context"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/hackathon"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *HackathonService) ValidateHackathon(ctx context.Context, req *hackathonv1.ValidateHackathonRequest) (*hackathonv1.ValidateHackathonResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	result, err := s.hackathonService.ValidateHackathon(ctx, hackathon.ValidateHackathonIn{
		HackathonID: hackathonID,
	})
	if err != nil {
		return nil, s.handleError(ctx, err, "validate_hackathon")
	}

	validationErrors := make([]*hackathonv1.ValidationError, 0, len(result.ValidationErrors))
	for _, ve := range result.ValidationErrors {
		validationErrors = append(validationErrors, &hackathonv1.ValidationError{
			Code:    ve.Code,
			Field:   ve.Field,
			Message: ve.Message,
			Meta:    ve.Meta,
		})
	}

	return &hackathonv1.ValidateHackathonResponse{
		ValidationErrors: validationErrors,
	}, nil
}
