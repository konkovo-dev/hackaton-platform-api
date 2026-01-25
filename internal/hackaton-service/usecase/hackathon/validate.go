package hackathon

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/google/uuid"
)

type ValidateHackathonIn struct {
	HackathonID uuid.UUID
}

type ValidateHackathonOut struct {
	ValidationErrors []domain.ValidationError
}

func (s *Service) ValidateHackathon(ctx context.Context, in ValidateHackathonIn) (*ValidateHackathonOut, error) {
	validatePolicy := hackathonpolicy.NewValidateHackathonPolicy(s.hackathonRepo, s.parClient)
	pctx, err := validatePolicy.LoadContext(ctx, hackathonpolicy.ValidateHackathonParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := validatePolicy.Check(ctx, pctx)
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
	validationErrors := validator.ValidateForPublish(hackathon, nil)

	return &ValidateHackathonOut{
		ValidationErrors: validationErrors,
	}, nil
}
