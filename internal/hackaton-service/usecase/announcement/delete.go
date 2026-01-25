package announcement

import (
	"context"

	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/google/uuid"
)

type DeleteAnnouncementIn struct {
	HackathonID    uuid.UUID
	AnnouncementID uuid.UUID
}

type DeleteAnnouncementOut struct{}

func (s *Service) DeleteAnnouncement(ctx context.Context, in DeleteAnnouncementIn) (*DeleteAnnouncementOut, error) {
	deletePolicy := hackathonpolicy.NewDeleteAnnouncementPolicy(s.hackathonRepo, s.parClient)
	pctx, err := deletePolicy.LoadContext(ctx, hackathonpolicy.AnnouncementPolicyParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := deletePolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	announcement, err := s.announcementRepo.GetByID(ctx, in.AnnouncementID)
	if err != nil {
		return nil, err
	}
	if announcement == nil {
		return nil, ErrAnnouncementNotFound
	}

	if err := s.announcementRepo.SoftDelete(ctx, in.AnnouncementID); err != nil {
		return nil, err
	}

	return &DeleteAnnouncementOut{}, nil
}
