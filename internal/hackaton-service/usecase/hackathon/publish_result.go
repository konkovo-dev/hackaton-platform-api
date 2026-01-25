package hackathon

import (
	"context"
	"time"

	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/google/uuid"
)

type PublishResultIn struct {
	HackathonID uuid.UUID
}

type PublishResultOut struct{}

func (s *Service) PublishResult(ctx context.Context, in PublishResultIn) (*PublishResultOut, error) {
	publishResultPolicy := hackathonpolicy.NewPublishResultPolicy(s.hackathonRepo, s.parClient)
	pctx, err := publishResultPolicy.LoadContext(ctx, hackathonpolicy.PublishResultParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := publishResultPolicy.Check(ctx, pctx)
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
	validationErrors := validator.ValidateResultReady(hackathon.Result)

	if len(validationErrors) > 0 {
		return nil, ErrValidationFailed
	}

	now := time.Now().UTC()
	hackathon.ResultPublishedAt = &now
	hackathon.JudgingEndsAt = &now

	hackathon.Stage = s.computeStage(now, hackathon)

	err = s.hackathonRepo.Update(ctx, hackathon)
	if err != nil {
		return nil, err
	}

	return &PublishResultOut{}, nil
}
