package hackathonservice

import (
	"context"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/hackathon"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *HackathonService) GetHackathonPermissions(ctx context.Context, req *hackathonv1.GetHackathonPermissionsRequest) (*hackathonv1.GetHackathonPermissionsResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid hackathon_id: %v", err)
	}

	in := hackathon.GetHackathonPermissionsIn{
		HackathonID: hackathonID,
	}

	out, err := s.hackathonService.GetHackathonPermissions(ctx, in)
	if err != nil {
		return nil, s.handleError(ctx, err, "GetHackathonPermissions")
	}

	return &hackathonv1.GetHackathonPermissionsResponse{
		Permissions: &hackathonv1.HackathonPermissions{
			ManageHackathon:    out.ManageHackathon,
			ReadDraft:          out.ReadDraft,
			PublishHackathon:   out.PublishHackathon,
			ViewAnnouncements:  out.ViewAnnouncements,
			CreateAnnouncement: out.CreateAnnouncement,
			ReadTask:           out.ReadTask,
			ViewResultPublic:   out.ViewResultPublic,
			ReadResultDraft:    out.ReadResultDraft,
			PublishResult:      out.PublishResult,
			UpdateResultDraft:  out.UpdateResultDraft,
		},
	}, nil
}
