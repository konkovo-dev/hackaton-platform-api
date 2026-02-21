package teamsservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/team"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) GetTeam(ctx context.Context, req *teamv1.GetTeamRequest) (*teamv1.GetTeamResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, team.ErrInvalidInput, "GetTeam")
	}

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, a.handleError(ctx, team.ErrInvalidInput, "GetTeam")
	}

	out, err := a.teamService.GetTeam(ctx, team.GetTeamIn{
		HackathonID:      hackathonID,
		TeamID:           teamID,
		IncludeVacancies: req.IncludeVacancies,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "GetTeam")
	}

	protoTeam := &teamv1.Team{
		TeamId:      out.Team.ID.String(),
		HackathonId: out.Team.HackathonID.String(),
		Name:        out.Team.Name,
		Description: out.Team.Description,
		IsJoinable:  out.Team.IsJoinable,
		CreatedAt:   timestamppb.New(out.Team.CreatedAt),
		UpdatedAt:   timestamppb.New(out.Team.UpdatedAt),
	}

	var protoVacancies []*teamv1.Vacancy
	if req.IncludeVacancies {
		protoVacancies = make([]*teamv1.Vacancy, 0, len(out.Vacancies))
		for _, v := range out.Vacancies {
			desiredRoleIDs := make([]string, len(v.DesiredRoleIDs))
			for i, id := range v.DesiredRoleIDs {
				desiredRoleIDs[i] = id.String()
			}

			desiredSkillIDs := make([]string, len(v.DesiredSkillIDs))
			for i, id := range v.DesiredSkillIDs {
				desiredSkillIDs[i] = id.String()
			}

			protoVacancies = append(protoVacancies, &teamv1.Vacancy{
				VacancyId:       v.ID.String(),
				TeamId:          v.TeamID.String(),
				Description:     v.Description,
				DesiredRoleIds:  desiredRoleIDs,
				DesiredSkillIds: desiredSkillIDs,
				SlotsTotal:      int64(v.SlotsTotal),
				SlotsOpen:       int64(v.SlotsOpen),
				CreatedAt:       timestamppb.New(v.CreatedAt),
				UpdatedAt:       timestamppb.New(v.UpdatedAt),
			})
		}
	}

	return &teamv1.GetTeamResponse{
		Team: &teamv1.TeamWithVacancies{
			Team:      protoTeam,
			Vacancies: protoVacancies,
		},
	}, nil
}

