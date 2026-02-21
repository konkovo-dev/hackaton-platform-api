package team

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	teampolicy "github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type GetTeamIn struct {
	HackathonID      uuid.UUID
	TeamID           uuid.UUID
	IncludeVacancies bool
}

type GetTeamOut struct {
	Team      *entity.Team
	Vacancies []*entity.Vacancy
}

func (s *Service) GetTeam(ctx context.Context, in GetTeamIn) (*GetTeamOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	_, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	stage, _, _, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	actorUserID, participationStatus, roles, err := s.parClient.GetHackathonContext(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon context: %w", err)
	}

	getPolicy := teampolicy.NewGetTeamPolicy()
	pctx, err := getPolicy.LoadContext(ctx, teampolicy.GetTeamParams{
		HackathonID: in.HackathonID,
		TeamID:      in.TeamID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)

	actorUUID, err := uuid.Parse(actorUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse actor user id: %w", err)
	}
	pctx.SetActorUserID(actorUUID)

	pctx.SetActorRoles(roles)
	pctx.SetParticipationStatus(participationStatus)
	pctx.SetHackathonStage(stage)

	decision := getPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	team, err := s.teamRepo.GetByIDAndHackathonID(ctx, in.TeamID, in.HackathonID)
	if err != nil {
		return nil, ErrNotFound
	}

	out := &GetTeamOut{
		Team: team,
	}

	if in.IncludeVacancies {
		vacancies, err := s.vacancyRepo.GetByTeamID(ctx, in.TeamID)
		if err != nil {
			return nil, fmt.Errorf("failed to get vacancies: %w", err)
		}
		out.Vacancies = vacancies
	}

	return out, nil
}

