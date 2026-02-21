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

type KickTeamMemberIn struct {
	HackathonID  uuid.UUID
	TeamID       uuid.UUID
	TargetUserID uuid.UUID
}

type KickTeamMemberOut struct{}

func (s *Service) KickTeamMember(ctx context.Context, in KickTeamMemberIn) (*KickTeamMemberOut, error) {
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

	isCaptain, err := s.membershipRepo.CheckIsCaptain(ctx, in.TeamID, userUUID)
	if err != nil {
		s.logger.Error("failed to check captain status", "error", err)
		return nil, fmt.Errorf("failed to check captain status: %w", err)
	}

	targetMembership, err := s.membershipRepo.GetByTeamAndUser(ctx, in.TeamID, in.TargetUserID)
	targetIsMember := err == nil
	targetIsCaptain := targetIsMember && targetMembership.IsCaptain

	kickPolicy := policy.NewKickTeamMemberPolicy()
	pctx, err := kickPolicy.LoadContext(ctx, policy.KickTeamMemberParams{
		HackathonID:  in.HackathonID,
		TeamID:       in.TeamID,
		TargetUserID: in.TargetUserID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetAllowTeam(allowTeam)
	pctx.SetIsCaptain(isCaptain)
	pctx.SetTargetIsMember(targetIsMember)
	pctx.SetTargetIsCaptain(targetIsCaptain)
	pctx.SetTargetUserID(in.TargetUserID)

	decision := kickPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		membershipRepoTx := txrepo.NewMembershipRepository(tx)
		if err := membershipRepoTx.Delete(ctx, in.TeamID, in.TargetUserID); err != nil {
			return fmt.Errorf("failed to delete membership: %w", err)
		}

		if targetMembership.AssignedVacancyID != nil {
			vacancyRepoTx := txrepo.NewVacancyRepository(tx)
			if err := vacancyRepoTx.IncrementSlotsOpen(ctx, *targetMembership.AssignedVacancyID); err != nil {
				return fmt.Errorf("failed to increment slots: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		s.logger.Error("failed to kick member transaction", "error", err)
		return nil, fmt.Errorf("failed to kick member: %w", err)
	}

	err = s.parClient.ConvertFromTeamParticipation(ctx, in.HackathonID.String(), in.TargetUserID.String())
	if err != nil {
		s.logger.Error("failed to convert from team participation, compensating", "error", err, "hackathon_id", in.HackathonID, "team_id", in.TeamID, "kicked_user_id", in.TargetUserID)
		
		compErr := s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
			membershipRepoTx := txrepo.NewMembershipRepository(tx)
			if err := membershipRepoTx.Create(ctx, targetMembership); err != nil {
				return err
			}
			
			if targetMembership.AssignedVacancyID != nil {
				vacancyRepoTx := txrepo.NewVacancyRepository(tx)
				if err := vacancyRepoTx.DecrementSlotsOpen(ctx, *targetMembership.AssignedVacancyID); err != nil {
					return err
				}
			}
			
			return nil
		})
		
		if compErr != nil {
			s.logger.Error("CRITICAL: failed to compensate kick member, data inconsistency", "error", compErr, "hackathon_id", in.HackathonID, "team_id", in.TeamID, "kicked_user_id", in.TargetUserID)
		}
		
		return nil, fmt.Errorf("failed to convert from team participation: %w", err)
	}

	return &KickTeamMemberOut{}, nil
}
