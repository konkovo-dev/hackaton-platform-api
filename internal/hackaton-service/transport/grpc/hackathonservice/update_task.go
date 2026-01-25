package hackathonservice

import (
	"context"
	"errors"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/hackathon"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *HackathonService) UpdateHackathonTask(ctx context.Context, req *hackathonv1.UpdateHackathonTaskRequest) (*hackathonv1.UpdateHackathonTaskResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid hackathon_id: %v", err)
	}

	in := hackathon.UpdateTaskIn{
		HackathonID: hackathonID,
		Task:        req.Task,
	}

	out, err := s.hackathonService.UpdateTask(ctx, in)
	if err != nil {
		if errors.Is(err, hackathon.ErrValidationFailed) && out != nil {
			validationErrors := make([]*hackathonv1.ValidationError, 0, len(out.ValidationErrors))
			for _, ve := range out.ValidationErrors {
				validationErrors = append(validationErrors, &hackathonv1.ValidationError{
					Code:    ve.Code,
					Field:   ve.Field,
					Message: ve.Message,
					Meta:    ve.Meta,
				})
			}
			return &hackathonv1.UpdateHackathonTaskResponse{
				ValidationErrors: validationErrors,
			}, nil
		}
		return nil, s.handleError(ctx, err, "UpdateHackathonTask")
	}

	validationErrors := make([]*hackathonv1.ValidationError, 0, len(out.ValidationErrors))
	for _, ve := range out.ValidationErrors {
		validationErrors = append(validationErrors, &hackathonv1.ValidationError{
			Code:    ve.Code,
			Field:   ve.Field,
			Message: ve.Message,
			Meta:    ve.Meta,
		})
	}

	return &hackathonv1.UpdateHackathonTaskResponse{
		ValidationErrors: validationErrors,
	}, nil
}
