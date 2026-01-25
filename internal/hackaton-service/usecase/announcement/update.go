package announcement

import (
	"context"

	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/google/uuid"
)

type UpdateAnnouncementIn struct {
	HackathonID    uuid.UUID
	AnnouncementID uuid.UUID
	Title          string
	Body           string
}

type UpdateAnnouncementOut struct{}

func (s *Service) UpdateAnnouncement(ctx context.Context, in UpdateAnnouncementIn) (*UpdateAnnouncementOut, error) {
	updatePolicy := hackathonpolicy.NewUpdateAnnouncementPolicy(s.hackathonRepo, s.parClient)
	pctx, err := updatePolicy.LoadContext(ctx, hackathonpolicy.AnnouncementPolicyParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := updatePolicy.Check(ctx, pctx)
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

	announcement.Title = in.Title
	announcement.Body = in.Body

	if err := s.announcementRepo.Update(ctx, announcement); err != nil {
		return nil, err
	}

	return &UpdateAnnouncementOut{}, nil
}
