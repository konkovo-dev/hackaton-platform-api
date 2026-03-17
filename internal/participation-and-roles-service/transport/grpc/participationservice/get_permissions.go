package participationservice

import (
	"context"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/participation"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *API) GetParticipationPermissions(ctx context.Context, req *participationrolesv1.GetParticipationPermissionsRequest) (*participationrolesv1.GetParticipationPermissionsResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid hackathon_id: %v", err)
	}

	in := participation.GetParticipationPermissionsIn{
		HackathonID: hackathonID,
	}

	out, err := a.participationService.GetParticipationPermissions(ctx, in)
	if err != nil {
		return nil, a.handleError(ctx, err, "GetParticipationPermissions")
	}

	return &participationrolesv1.GetParticipationPermissionsResponse{
		Permissions: &participationrolesv1.ParticipationPermissions{
			Register:                   out.Register,
			Unregister:                 out.Unregister,
			SwitchParticipationMode:    out.SwitchParticipationMode,
			UpdateParticipationProfile: out.UpdateParticipationProfile,
			InviteStaff:                out.InviteStaff,
			ListParticipants:           out.ListParticipants,
		},
	}, nil
}
