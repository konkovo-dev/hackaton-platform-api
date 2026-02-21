package teamsservice

import (
	"context"
	"log/slog"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/team"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) ListTeams(ctx context.Context, req *teamv1.ListTeamsRequest) (*teamv1.ListTeamsResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	var pageSize uint32
	var pageToken string

	if req.Query != nil && req.Query.Page != nil {
		pageSize = req.Query.Page.PageSize
		pageToken = req.Query.Page.PageToken
	}

	result, err := a.teamService.ListTeams(ctx, team.ListTeamsIn{
		HackathonID:      hackathonID,
		PageSize:         pageSize,
		PageToken:        pageToken,
		IncludeVacancies: req.IncludeVacancies,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "list_teams")
	}

	protoTeams := make([]*teamv1.TeamWithVacancies, 0, len(result.Teams))
	for _, t := range result.Teams {
		protoTeam := &teamv1.Team{
			TeamId:      t.ID.String(),
			HackathonId: t.HackathonID.String(),
			Name:        t.Name,
			Description: t.Description,
			IsJoinable:  t.IsJoinable,
			CreatedAt:   timestamppb.New(t.CreatedAt),
			UpdatedAt:   timestamppb.New(t.UpdatedAt),
		}

		protoVacancies := []*teamv1.Vacancy{}
		if req.IncludeVacancies {
			if vacancies, ok := result.Vacancies[t.ID.String()]; ok {
				for _, v := range vacancies {
					roleIDs := make([]string, len(v.DesiredRoleIDs))
					for i, id := range v.DesiredRoleIDs {
						roleIDs[i] = id.String()
					}

					skillIDs := make([]string, len(v.DesiredSkillIDs))
					for i, id := range v.DesiredSkillIDs {
						skillIDs[i] = id.String()
					}

					protoVacancies = append(protoVacancies, &teamv1.Vacancy{
						VacancyId:        v.ID.String(),
						TeamId:           v.TeamID.String(),
						Description:      v.Description,
						DesiredRoleIds:   roleIDs,
						DesiredSkillIds:  skillIDs,
						SlotsTotal:       int64(v.SlotsTotal),
						SlotsOpen:        int64(v.SlotsOpen),
						CreatedAt:        timestamppb.New(v.CreatedAt),
						UpdatedAt:        timestamppb.New(v.UpdatedAt),
					})
				}
			}
		}

		protoTeams = append(protoTeams, &teamv1.TeamWithVacancies{
			Team:      protoTeam,
			Vacancies: protoVacancies,
		})
	}

	a.logger.InfoContext(ctx, "list_teams: success",
		slog.String("hackathon_id", req.HackathonId),
		slog.Int("count", len(protoTeams)),
	)

	return &teamv1.ListTeamsResponse{
		Teams: protoTeams,
		Page: &commonv1.PageResponse{
			NextPageToken: result.NextPageToken,
		},
	}, nil
}
