package team

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	teampolicy "github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/repository/postgres"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CreateTeamIn struct {
	HackathonID uuid.UUID
	Name        string
	Description string
	IsJoinable  bool
}

type CreateTeamOut struct {
	TeamID uuid.UUID
}

func (s *Service) CreateTeam(ctx context.Context, in CreateTeamIn) (*CreateTeamOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if in.Name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	stage, allowTeam, _, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	actorUserID, participationStatus, roles, err := s.parClient.GetHackathonContext(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon context: %w", err)
	}

	createPolicy := teampolicy.NewCreateTeamPolicy()
	pctx, err := createPolicy.LoadContext(ctx, teampolicy.CreateTeamParams{
		HackathonID: in.HackathonID,
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
	pctx.SetAllowTeam(allowTeam)

	decision := createPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	exists, err := s.teamRepo.CheckNameExists(ctx, in.HackathonID, in.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check team name: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("%w: team name already exists in this hackathon", ErrConflict)
	}

	teamID := uuid.New()
	now := time.Now().UTC()

	team := &entity.Team{
		ID:          teamID,
		HackathonID: in.HackathonID,
		Name:        in.Name,
		Description: in.Description,
		IsJoinable:  in.IsJoinable,
	}

	membership := &entity.Membership{
		TeamID:            teamID,
		UserID:            userUUID,
		IsCaptain:         true,
		AssignedVacancyID: nil,
		JoinedAt:          now,
	}

	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		teamRepoTx := NewTeamRepositoryWithTx(tx)
		if err := teamRepoTx.Create(ctx, team); err != nil {
			return fmt.Errorf("failed to create team: %w", err)
		}

		membershipRepoTx := NewMembershipRepositoryWithTx(tx)
		if err := membershipRepoTx.Create(ctx, membership); err != nil {
			return fmt.Errorf("failed to create membership: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	err = s.parClient.ConvertToTeamParticipation(ctx, in.HackathonID.String(), userUUID.String(), teamID.String(), true)
	if err != nil {
		deleteErr := s.teamRepo.Delete(ctx, teamID)
		if deleteErr != nil {
			s.logger.Error("failed to compensate team creation", "team_id", teamID.String(), "error", deleteErr)
		}
		return nil, fmt.Errorf("failed to convert participation: %w", err)
	}

	return &CreateTeamOut{
		TeamID: teamID,
	}, nil
}

func NewTeamRepositoryWithTx(tx pgx.Tx) TeamRepository {
	return postgres.NewTeamRepository(tx)
}

func NewMembershipRepositoryWithTx(tx pgx.Tx) MembershipRepository {
	return postgres.NewMembershipRepository(tx)
}
