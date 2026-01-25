package hackathon

import (
	"context"

	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/google/uuid"
)

type GetTaskIn struct {
	HackathonID uuid.UUID
}

type GetTaskOut struct {
	Task string
}

func (s *Service) GetTask(ctx context.Context, in GetTaskIn) (*GetTaskOut, error) {
	readTaskPolicy := hackathonpolicy.NewReadTaskPolicy(s.hackathonRepo, s.parClient)
	pctx, err := readTaskPolicy.LoadContext(ctx, hackathonpolicy.ReadTaskParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := readTaskPolicy.Check(ctx, pctx)
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

	return &GetTaskOut{
		Task: hackathon.Task,
	}, nil
}
