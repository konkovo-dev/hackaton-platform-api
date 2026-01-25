package announcement

import (
	"context"

	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/domain/entity"
	hackathonpolicy "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type ListAnnouncementsIn struct {
	HackathonID uuid.UUID
}

type ListAnnouncementsOut struct {
	Announcements []*entity.HackathonAnnouncement
}

func (s *Service) ListAnnouncements(ctx context.Context, in ListAnnouncementsIn) (*ListAnnouncementsOut, error) {
	readPolicy := hackathonpolicy.NewReadAnnouncementsPolicy(s.hackathonRepo, s.parClient)
	pctx, err := readPolicy.LoadContext(ctx, hackathonpolicy.AnnouncementPolicyParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := readPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	announcements, err := s.announcementRepo.ListByHackathonID(ctx, in.HackathonID)
	if err != nil {
		return nil, err
	}

	return &ListAnnouncementsOut{
		Announcements: announcements,
	}, nil
}

func mapPolicyError(decision *policy.Decision) error {
	if len(decision.Violations) == 0 {
		return ErrUnauthorized
	}

	v := decision.Violations[0]
	if v.Code == policy.ViolationCodeForbidden {
		return ErrForbidden
	}

	return ErrUnauthorized
}
