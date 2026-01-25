package announcement

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/google/uuid"
)

type CreateAnnouncementIn struct {
	HackathonID uuid.UUID
	Title       string
	Body        string
}

type CreateAnnouncementOut struct {
	AnnouncementID uuid.UUID
}

func (s *Service) CreateAnnouncement(ctx context.Context, in CreateAnnouncementIn) (*CreateAnnouncementOut, error) {
	createPolicy := hackathonpolicy.NewCreateAnnouncementPolicy(s.hackathonRepo, s.parClient)
	pctx, err := createPolicy.LoadContext(ctx, hackathonpolicy.AnnouncementPolicyParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := createPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	hctx := pctx.(*hackathonpolicy.HackathonPolicyContext)
	creatorUserID := hctx.ActorUserID()

	announcementID := uuid.New()
	announcement := &entity.HackathonAnnouncement{
		ID:            announcementID,
		HackathonID:   in.HackathonID,
		Title:         in.Title,
		Body:          in.Body,
		CreatedByUser: creatorUserID,
	}

	if err := s.announcementRepo.Create(ctx, announcement); err != nil {
		return nil, err
	}

	return &CreateAnnouncementOut{
		AnnouncementID: announcementID,
	}, nil
}
