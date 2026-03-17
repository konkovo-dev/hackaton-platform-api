package hackathon

import (
	"context"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/google/uuid"
)

type PublishHackathonIn struct {
	HackathonID uuid.UUID
}

type PublishHackathonOut struct {
	PublishedAt time.Time
}

func (s *Service) PublishHackathon(ctx context.Context, in PublishHackathonIn) (*PublishHackathonOut, error) {
	publishPolicy := hackathonpolicy.NewPublishHackathonPolicy(s.hackathonRepo, s.parClient)
	pctx, err := publishPolicy.LoadContext(ctx, hackathonpolicy.PublishHackathonParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := publishPolicy.Check(ctx, pctx)
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
	if len(validationErrors) > 0 {
		return nil, ErrValidationFailed
	}

	now := time.Now().UTC()

	if hackathon.RegistrationOpensAt != nil && !now.Before(*hackathon.RegistrationOpensAt) {
		return nil, ErrValidationFailed
	}

	hackathon.PublishedAt = &now
	hackathon.State = string(domain.StatePublished)
	hackathon.Stage = domain.ComputeStage(now, hackathon)

	err = s.uow.Do(ctx, func(ctx context.Context, txRepos *TxRepositories) error {
		return txRepos.Hackathons.Update(ctx, hackathon)
	})

	if err != nil {
		return nil, err
	}

	return &PublishHackathonOut{
		PublishedAt: *hackathon.PublishedAt,
	}, nil
}

