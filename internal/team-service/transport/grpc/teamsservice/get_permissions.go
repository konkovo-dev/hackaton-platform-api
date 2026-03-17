package teamsservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/team"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *API) GetTeamPermissions(ctx context.Context, req *teamv1.GetTeamPermissionsRequest) (*teamv1.GetTeamPermissionsResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid hackathon_id: %v", err)
	}

	in := team.GetTeamPermissionsIn{
		HackathonID: hackathonID,
	}

	out, err := a.teamService.GetTeamPermissions(ctx, in)
	if err != nil {
		return nil, a.handleError(ctx, err, "GetTeamPermissions")
	}

	return &teamv1.GetTeamPermissionsResponse{
		Permissions: &teamv1.TeamPermissions{
			CreateTeam: out.CreateTeam,
			CanInMyTeam: &teamv1.TeamManagementPermissions{
				EditTeam:           out.CanInMyTeam.EditTeam,
				DeleteTeam:         out.CanInMyTeam.DeleteTeam,
				ManageVacancies:    out.CanInMyTeam.ManageVacancies,
				InviteMember:       out.CanInMyTeam.InviteMember,
				ManageJoinRequests: out.CanInMyTeam.ManageJoinRequests,
				KickMember:         out.CanInMyTeam.KickMember,
				TransferCaptain:    out.CanInMyTeam.TransferCaptain,
				LeaveTeam:          out.CanInMyTeam.LeaveTeam,
			},
		},
	}, nil
}
