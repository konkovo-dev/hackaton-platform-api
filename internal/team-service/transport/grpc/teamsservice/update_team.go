package teamsservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/team"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) UpdateTeam(ctx context.Context, req *teamv1.UpdateTeamRequest) (*teamv1.UpdateTeamResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, team.ErrInvalidInput, "UpdateTeam")
	}

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, a.handleError(ctx, team.ErrInvalidInput, "UpdateTeam")
	}

	out, err := a.teamService.UpdateTeam(ctx, team.UpdateTeamIn{
		HackathonID: hackathonID,
		TeamID:      teamID,
		Name:        req.Name,
		Description: req.Description,
		IsJoinable:  req.IsJoinable,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "UpdateTeam")
	}

	return &teamv1.UpdateTeamResponse{
		Team: &teamv1.Team{
			TeamId:      out.Team.ID.String(),
			HackathonId: out.Team.HackathonID.String(),
			Name:        out.Team.Name,
			Description: out.Team.Description,
			IsJoinable:  out.Team.IsJoinable,
			CreatedAt:   timestamppb.New(out.Team.CreatedAt),
			UpdatedAt:   timestamppb.New(out.Team.UpdatedAt),
		},
	}, nil
}
