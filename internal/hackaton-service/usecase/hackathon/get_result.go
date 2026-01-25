package hackathon

import (
	"context"

	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/google/uuid"
)

type GetResultIn struct {
	HackathonID uuid.UUID
}

type GetResultOut struct {
	Result string
}

func (s *Service) GetResult(ctx context.Context, in GetResultIn) (*GetResultOut, error) {
	readResultPolicy := hackathonpolicy.NewReadResultPolicy(s.hackathonRepo, s.parClient)
	pctx, err := readResultPolicy.LoadContext(ctx, hackathonpolicy.ReadResultParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := readResultPolicy.Check(ctx, pctx)
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

	return &GetResultOut{
		Result: hackathon.Result,
	}, nil
}
