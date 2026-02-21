package teammember

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/txrepo"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type LeaveTeamIn struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
}

type LeaveTeamOut struct{}

func (s *Service) LeaveTeam(ctx context.Context, in LeaveTeamIn) (*LeaveTeamOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	stage, allowTeam, _, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		s.logger.Error("failed to get hackathon", "error", err)
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	_, err = s.teamRepo.GetByIDAndHackathonID(ctx, in.TeamID, in.HackathonID)
	if err != nil {
		return nil, ErrNotFound
	}

	membership, err := s.membershipRepo.GetByTeamAndUser(ctx, in.TeamID, userUUID)
	isMember := err == nil
	isCaptain := isMember && membership.IsCaptain

	leavePolicy := policy.NewLeaveTeamPolicy()
	pctx, err := leavePolicy.LoadContext(ctx, policy.LeaveTeamParams{
		HackathonID: in.HackathonID,
		TeamID:      in.TeamID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetAllowTeam(allowTeam)
	pctx.SetIsMember(isMember)
	pctx.SetIsCaptain(isCaptain)

	decision := leavePolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		membershipRepoTx := txrepo.NewMembershipRepository(tx)
		if err := membershipRepoTx.Delete(ctx, in.TeamID, userUUID); err != nil {
			return fmt.Errorf("failed to delete membership: %w", err)
		}

		if membership.AssignedVacancyID != nil {
			vacancyRepoTx := txrepo.NewVacancyRepository(tx)
			if err := vacancyRepoTx.IncrementSlotsOpen(ctx, *membership.AssignedVacancyID); err != nil {
				return fmt.Errorf("failed to increment slots: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		s.logger.Error("failed to leave team transaction", "error", err)
		return nil, fmt.Errorf("failed to leave team: %w", err)
	}

	err = s.parClient.ConvertFromTeamParticipation(ctx, in.HackathonID.String(), userUUID.String())
	if err != nil {
		s.logger.Error("failed to convert from team participation, compensating", "error", err, "hackathon_id", in.HackathonID, "team_id", in.TeamID, "user_id", userUUID)
		
		compErr := s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
			membershipRepoTx := txrepo.NewMembershipRepository(tx)
			if err := membershipRepoTx.Create(ctx, membership); err != nil {
				return err
			}
			
			if membership.AssignedVacancyID != nil {
				vacancyRepoTx := txrepo.NewVacancyRepository(tx)
				if err := vacancyRepoTx.DecrementSlotsOpen(ctx, *membership.AssignedVacancyID); err != nil {
					return err
				}
			}
			
			return nil
		})
		
		if compErr != nil {
			s.logger.Error("CRITICAL: failed to compensate leave team, data inconsistency", "error", compErr, "hackathon_id", in.HackathonID, "team_id", in.TeamID, "user_id", userUUID)
		}
		
		return nil, fmt.Errorf("failed to convert from team participation: %w", err)
	}

	return &LeaveTeamOut{}, nil
}
