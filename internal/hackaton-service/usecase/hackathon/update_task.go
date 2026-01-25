package hackathon

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/google/uuid"
)

type UpdateTaskIn struct {
	HackathonID uuid.UUID
	Task        string
}

type UpdateTaskOut struct {
	ValidationErrors []domain.ValidationError
}

func (s *Service) UpdateTask(ctx context.Context, in UpdateTaskIn) (*UpdateTaskOut, error) {
	updateTaskPolicy := hackathonpolicy.NewUpdateTaskPolicy(s.hackathonRepo, s.parClient)
	pctx, err := updateTaskPolicy.LoadContext(ctx, hackathonpolicy.UpdateTaskParams{
		HackathonID: in.HackathonID,
		NewTask:     in.Task,
	})
	if err != nil {
		return nil, err
	}

	decision := updateTaskPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, s.mapPolicyError(decision)
	}

	hackathon, err := s.hackathonRepo.GetByID(ctx, in.HackathonID)
	if err != nil {
		return nil, err
	}
	if hackathon == nil {
		return nil, ErrHackathonNotFound
	}

	validator := NewHackathonValidator()
	stage := domain.HackathonStage(hackathon.Stage)
	validationErrors := validator.ValidateTaskUpdate(in.Task, stage)

	if stage != domain.StageDraft && len(validationErrors) > 0 {
		return &UpdateTaskOut{
			ValidationErrors: validationErrors,
		}, ErrValidationFailed
	}

	hackathon.Task = in.Task

	err = s.hackathonRepo.Update(ctx, hackathon)
	if err != nil {
		return nil, err
	}

	return &UpdateTaskOut{
		ValidationErrors: validationErrors,
	}, nil
}
