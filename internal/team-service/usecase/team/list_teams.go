package team

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	teampolicy "github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

const (
	DefaultPageSize = 50
	MaxPageSize     = 100
)

type ListTeamsIn struct {
	HackathonID       uuid.UUID
	PageSize          uint32
	PageToken         string
	IncludeVacancies  bool
}

type ListTeamsOut struct {
	Teams         []*entity.Team
	Vacancies     map[string][]*entity.Vacancy
	NextPageToken string
}

func (s *Service) ListTeams(ctx context.Context, in ListTeamsIn) (*ListTeamsOut, error) {
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

	listPolicy := teampolicy.NewListTeamsPolicy()
	pctx, err := listPolicy.LoadContext(ctx, teampolicy.ListTeamsParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)

	actorUUID, err := uuid.Parse(actorUserID)
	if err != nil {
		return nil, ErrUnauthorized
	}
	pctx.SetActorUserID(actorUUID)
	pctx.SetActorRoles(roles)
	pctx.SetParticipationStatus(participationStatus)
	pctx.SetHackathonStage(stage)

	decision := listPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	pageSize := in.PageSize
	if pageSize == 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	offset := uint32(0)
	if in.PageToken != "" {
		parsedOffset, err := parsePageToken(in.PageToken)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid page token", ErrInvalidInput)
		}
		offset = parsedOffset
	}

	teams, err := s.teamRepo.List(ctx, in.HackathonID, int32(pageSize), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}

	nextPageToken := ""
	if len(teams) == int(pageSize) {
		nextPageToken = encodePageToken(offset + pageSize)
	}

	vacanciesMap := make(map[string][]*entity.Vacancy)
	if in.IncludeVacancies && len(teams) > 0 {
		teamIDs := make([]uuid.UUID, len(teams))
		for i, t := range teams {
			teamIDs[i] = t.ID
		}

		allVacancies, err := s.vacancyRepo.GetByTeamIDs(ctx, teamIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get vacancies: %w", err)
		}

		for _, vacancy := range allVacancies {
			vacanciesMap[vacancy.TeamID.String()] = append(vacanciesMap[vacancy.TeamID.String()], vacancy)
		}
	}

	return &ListTeamsOut{
		Teams:         teams,
		Vacancies:     vacanciesMap,
		NextPageToken: nextPageToken,
	}, nil
}

func mapPolicyError(decision *policy.Decision) error {
	if len(decision.Violations) == 0 {
		return ErrUnauthorized
	}

	v := decision.Violations[0]
	switch v.Code {
	case policy.ViolationCodeForbidden:
		return fmt.Errorf("%w: %s", ErrForbidden, v.Message)
	case policy.ViolationCodeNotFound:
		return fmt.Errorf("%w: %s", ErrNotFound, v.Message)
	case policy.ViolationCodeStageRule, policy.ViolationCodeTimeRule, policy.ViolationCodePolicyRule:
		return fmt.Errorf("%w: %s", ErrForbidden, v.Message)
	case policy.ViolationCodeConflict:
		return fmt.Errorf("%w: %s", ErrConflict, v.Message)
	default:
		return ErrUnauthorized
	}
}
