package teamsservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/team"
	"github.com/google/uuid"
)

func (a *API) CreateTeam(ctx context.Context, req *teamv1.CreateTeamRequest) (*teamv1.CreateTeamResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, team.ErrInvalidInput, "CreateTeam")
	}

	out, err := a.teamService.CreateTeam(ctx, team.CreateTeamIn{
		HackathonID: hackathonID,
		Name:        req.Name,
		Description: req.Description,
		IsJoinable:  req.IsJoinable,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "CreateTeam")
	}

	return &teamv1.CreateTeamResponse{
		TeamId: out.TeamID.String(),
	}, nil
}
