package teammembersservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teammember"
	"github.com/google/uuid"
)

func (a *API) LeaveTeam(ctx context.Context, req *teamv1.LeaveTeamRequest) (*teamv1.LeaveTeamResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, err, "LeaveTeam")
	}

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, a.handleError(ctx, err, "LeaveTeam")
	}

	_, err = a.teamMemberService.LeaveTeam(ctx, teammember.LeaveTeamIn{
		HackathonID: hackathonID,
		TeamID:      teamID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "LeaveTeam")
	}

	return &teamv1.LeaveTeamResponse{}, nil
}
