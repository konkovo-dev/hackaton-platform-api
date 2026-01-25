package hackathonservice

import (
	"context"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/announcement"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *HackathonService) ListHackathonAnnouncements(ctx context.Context, req *hackathonv1.ListHackathonAnnouncementsRequest) (*hackathonv1.ListHackathonAnnouncementsResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	result, err := s.announcementService.ListAnnouncements(ctx, announcement.ListAnnouncementsIn{
		HackathonID: hackathonID,
	})
	if err != nil {
		return nil, s.handleAnnouncementError(ctx, err, "list_announcements")
	}

	announcements := make([]*hackathonv1.HackathonAnnouncement, 0, len(result.Announcements))
	for _, a := range result.Announcements {
		announcements = append(announcements, &hackathonv1.HackathonAnnouncement{
			AnnouncementId: a.ID.String(),
			Title:          a.Title,
			Body:           a.Body,
			CreatedAt:      timestamppb.New(a.CreatedAt),
			UpdatedAt:      timestamppb.New(a.UpdatedAt),
		})
	}

	return &hackathonv1.ListHackathonAnnouncementsResponse{
		Announcements: announcements,
	}, nil
}
