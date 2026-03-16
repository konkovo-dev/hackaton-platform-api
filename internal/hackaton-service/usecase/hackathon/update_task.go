package hackathon

import (
	"context"

	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/google/uuid"
)

type UpdateTaskIn struct {
	HackathonID uuid.UUID
	Task        string
}

type UpdateTaskOut struct {
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

	hackathon.Task = in.Task

	err = s.hackathonRepo.Update(ctx, hackathon)
	if err != nil {
		return nil, err
	}

	return &UpdateTaskOut{}, nil
}
