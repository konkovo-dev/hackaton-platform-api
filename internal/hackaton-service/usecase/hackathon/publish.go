package hackathon

import (
	"context"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
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
	hackathon.Stage = s.computeStage(now, hackathon)

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

func (s *Service) computeStage(now time.Time, hackathon *entity.Hackathon) string {
	if hackathon.PublishedAt == nil {
		return string(domain.StageDraft)
	}

	if hackathon.RegistrationOpensAt != nil && now.Before(*hackathon.RegistrationOpensAt) {
		return string(domain.StageUpcoming)
	}
	if hackathon.RegistrationClosesAt != nil && now.Before(*hackathon.RegistrationClosesAt) {
		return string(domain.StageRegistration)
	}
	if hackathon.StartsAt != nil && now.Before(*hackathon.StartsAt) {
		return string(domain.StagePreStart)
	}
	if hackathon.EndsAt != nil && now.Before(*hackathon.EndsAt) {
		return string(domain.StageRunning)
	}
	if hackathon.JudgingEndsAt != nil && now.Before(*hackathon.JudgingEndsAt) && hackathon.ResultPublishedAt == nil {
		return string(domain.StageJudging)
	}
	return string(domain.StageFinished)
}
