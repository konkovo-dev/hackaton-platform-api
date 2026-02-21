package teamsservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/team"
	"github.com/google/uuid"
)

func (a *API) DeleteTeam(ctx context.Context, req *teamv1.DeleteTeamRequest) (*teamv1.DeleteTeamResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, err, "DeleteTeam")
	}

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, a.handleError(ctx, err, "DeleteTeam")
	}

	_, err = a.teamService.DeleteTeam(ctx, team.DeleteTeamIn{
		HackathonID: hackathonID,
		TeamID:      teamID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "DeleteTeam")
	}

	return &teamv1.DeleteTeamResponse{}, nil
}
