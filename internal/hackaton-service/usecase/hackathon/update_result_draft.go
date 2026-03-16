package hackathon

import (
	"context"

	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/google/uuid"
)

type UpdateResultDraftIn struct {
	HackathonID uuid.UUID
	Result      string
}

type UpdateResultDraftOut struct {
}

func (s *Service) UpdateResultDraft(ctx context.Context, in UpdateResultDraftIn) (*UpdateResultDraftOut, error) {
	updateResultPolicy := hackathonpolicy.NewUpdateResultDraftPolicy(s.hackathonRepo, s.parClient)
	pctx, err := updateResultPolicy.LoadContext(ctx, hackathonpolicy.UpdateResultDraftParams{
		HackathonID: in.HackathonID,
		NewResult:   in.Result,
	})
	if err != nil {
		return nil, err
	}

	decision := updateResultPolicy.Check(ctx, pctx)
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

	hackathon.Result = in.Result

	err = s.hackathonRepo.Update(ctx, hackathon)
	if err != nil {
		return nil, err
	}

	return &UpdateResultDraftOut{}, nil
}
