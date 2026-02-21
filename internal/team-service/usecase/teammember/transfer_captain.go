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

type TransferCaptainIn struct {
	HackathonID  uuid.UUID
	TeamID       uuid.UUID
	NewCaptainID uuid.UUID
}

type TransferCaptainOut struct{}

func (s *Service) TransferCaptain(ctx context.Context, in TransferCaptainIn) (*TransferCaptainOut, error) {
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

	_, err = s.membershipRepo.GetByTeamAndUser(ctx, in.TeamID, in.NewCaptainID)
	targetIsMember := err == nil

	transferPolicy := policy.NewTransferCaptainPolicy()
	pctx, err := transferPolicy.LoadContext(ctx, policy.TransferCaptainParams{
		HackathonID:  in.HackathonID,
		TeamID:       in.TeamID,
		NewCaptainID: in.NewCaptainID,
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
	pctx.SetNewCaptainID(in.NewCaptainID)

	decision := transferPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		membershipRepoTx := txrepo.NewMembershipRepository(tx)

		if err := membershipRepoTx.UpdateCaptainStatus(ctx, in.TeamID, userUUID, false); err != nil {
			return fmt.Errorf("failed to demote old captain: %w", err)
		}

		if err := membershipRepoTx.UpdateCaptainStatus(ctx, in.TeamID, in.NewCaptainID, true); err != nil {
			return fmt.Errorf("failed to promote new captain: %w", err)
		}

		return nil
	})
	if err != nil {
		s.logger.Error("failed to transfer captain transaction", "error", err)
		return nil, fmt.Errorf("failed to transfer captain: %w", err)
	}

	if err := s.parClient.ConvertToTeamParticipation(ctx, in.HackathonID.String(), userUUID.String(), in.TeamID.String(), false); err != nil {
		s.logger.Error("failed to convert old captain to member, compensating", "error", err, "hackathon_id", in.HackathonID, "team_id", in.TeamID)
		
		compErr := s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
			membershipRepoTx := txrepo.NewMembershipRepository(tx)
			if err := membershipRepoTx.UpdateCaptainStatus(ctx, in.TeamID, userUUID, true); err != nil {
				return err
			}
			if err := membershipRepoTx.UpdateCaptainStatus(ctx, in.TeamID, in.NewCaptainID, false); err != nil {
				return err
			}
			return nil
		})
		
		if compErr != nil {
			s.logger.Error("CRITICAL: failed to compensate captain transfer, data inconsistency", "error", compErr, "hackathon_id", in.HackathonID, "team_id", in.TeamID, "old_captain", userUUID, "new_captain", in.NewCaptainID)
		}
		
		return nil, fmt.Errorf("failed to convert old captain to member: %w", err)
	}

	if err := s.parClient.ConvertToTeamParticipation(ctx, in.HackathonID.String(), in.NewCaptainID.String(), in.TeamID.String(), true); err != nil {
		s.logger.Error("failed to convert new captain, compensating", "error", err, "hackathon_id", in.HackathonID, "team_id", in.TeamID)
		
		if compErr := s.parClient.ConvertToTeamParticipation(ctx, in.HackathonID.String(), userUUID.String(), in.TeamID.String(), true); compErr != nil {
			s.logger.Error("CRITICAL: failed to restore old captain participation status, data inconsistency", "error", compErr, "hackathon_id", in.HackathonID, "team_id", in.TeamID, "old_captain", userUUID)
		}
		
		compErr := s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
			membershipRepoTx := txrepo.NewMembershipRepository(tx)
			if err := membershipRepoTx.UpdateCaptainStatus(ctx, in.TeamID, userUUID, true); err != nil {
				return err
			}
			if err := membershipRepoTx.UpdateCaptainStatus(ctx, in.TeamID, in.NewCaptainID, false); err != nil {
				return err
			}
			return nil
		})
		
		if compErr != nil {
			s.logger.Error("CRITICAL: failed to compensate captain transfer in DB, data inconsistency", "error", compErr, "hackathon_id", in.HackathonID, "team_id", in.TeamID, "old_captain", userUUID, "new_captain", in.NewCaptainID)
		}
		
		return nil, fmt.Errorf("failed to convert new captain: %w", err)
	}

	return &TransferCaptainOut{}, nil
}
