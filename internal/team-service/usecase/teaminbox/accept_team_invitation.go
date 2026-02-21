package teaminbox

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/repository/postgres"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AcceptTeamInvitationIn struct {
	InvitationID uuid.UUID
}

type AcceptTeamInvitationOut struct{}

func (s *Service) AcceptTeamInvitation(ctx context.Context, in AcceptTeamInvitationIn) (*AcceptTeamInvitationOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	invitation, err := s.invitationRepo.GetByID(ctx, in.InvitationID)
	if err != nil {
		return nil, ErrNotFound
	}

	stage, allowTeam, _, err := s.hackathonClient.GetHackathon(ctx, invitation.HackathonID.String())
	if err != nil {
		s.logger.Error("failed to get hackathon", "error", err)
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	_, participationStatus, roles, err := s.parClient.GetHackathonContext(ctx, invitation.HackathonID.String())
	if err != nil {
		s.logger.Error("failed to get hackathon context", "error", err)
		return nil, fmt.Errorf("failed to get hackathon context: %w", err)
	}

	isStaff := false
	for _, role := range roles {
		role = strings.ToLower(role)
		if role == "owner" || role == "organizer" || role == "mentor" {
			isStaff = true
			break
		}
	}

	vacancy, err := s.vacancyRepo.GetByID(ctx, invitation.VacancyID)
	if err != nil {
		return nil, fmt.Errorf("%w: vacancy not found", ErrNotFound)
	}

	if vacancy.SlotsOpen <= 0 {
		return nil, fmt.Errorf("%w: no open slots in vacancy", ErrConflict)
	}

	acceptPolicy := policy.NewAcceptTeamInvitationPolicy()
	pctx, err := acceptPolicy.LoadContext(ctx, policy.AcceptTeamInvitationParams{
		InvitationID: in.InvitationID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetAllowTeam(allowTeam)
	pctx.SetInvitationStatus(invitation.Status)
	pctx.SetInvitationTarget(invitation.TargetUserID)
	pctx.SetIsStaff(isStaff)
	pctx.SetParticipationStatus(participationStatus)

	decision := acceptPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	now := time.Now().UTC()
	membership := &entity.Membership{
		TeamID:            invitation.TeamID,
		UserID:            userUUID,
		IsCaptain:         false,
		AssignedVacancyID: &invitation.VacancyID,
		JoinedAt:          now,
	}

	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		invitationRepoTx := postgres.NewTeamInvitationRepository(tx)
		if err := invitationRepoTx.UpdateStatus(ctx, in.InvitationID, "accepted"); err != nil {
			return fmt.Errorf("failed to accept invitation: %w", err)
		}

		membershipRepoTx := postgres.NewMembershipRepository(tx)
		if err := membershipRepoTx.Create(ctx, membership); err != nil {
			return fmt.Errorf("failed to create membership: %w", err)
		}

		vacancyRepoTx := postgres.NewVacancyRepository(tx)
		if err := vacancyRepoTx.DecrementSlotsOpen(ctx, invitation.VacancyID); err != nil {
			return fmt.Errorf("failed to decrement slots: %w", err)
		}

		if err := invitationRepoTx.CancelCompeting(ctx, userUUID, invitation.HackathonID); err != nil {
			return fmt.Errorf("failed to cancel competing invitations: %w", err)
		}

		joinRequestRepoTx := postgres.NewJoinRequestRepository(tx)
		if err := joinRequestRepoTx.CancelCompeting(ctx, userUUID, invitation.HackathonID); err != nil {
			return fmt.Errorf("failed to cancel competing join requests: %w", err)
		}

		return nil
	})
	if err != nil {
		s.logger.Error("failed to accept invitation transaction", "error", err)
		return nil, fmt.Errorf("failed to accept invitation: %w", err)
	}

	err = s.parClient.ConvertToTeamParticipation(ctx, invitation.HackathonID.String(), userUUID.String(), invitation.TeamID.String(), false)
	if err != nil {
		s.logger.Error("failed to convert to team participation, compensating", "error", err, "hackathon_id", invitation.HackathonID, "team_id", invitation.TeamID, "user_id", userUUID)
		
		compErr := s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
			invRepoTx := postgres.NewTeamInvitationRepository(tx)
			if err := invRepoTx.UpdateStatus(ctx, in.InvitationID, "pending"); err != nil {
				return err
			}
			
			membershipRepoTx := postgres.NewMembershipRepository(tx)
			if err := membershipRepoTx.Delete(ctx, invitation.TeamID, userUUID); err != nil {
				return err
			}
			
			vacancyRepoTx := postgres.NewVacancyRepository(tx)
			if err := vacancyRepoTx.IncrementSlotsOpen(ctx, invitation.VacancyID); err != nil {
				return err
			}
			
			return nil
		})
		
		if compErr != nil {
			s.logger.Error("CRITICAL: failed to compensate invitation acceptance, data inconsistency", "error", compErr, "hackathon_id", invitation.HackathonID, "team_id", invitation.TeamID, "user_id", userUUID, "invitation_id", in.InvitationID)
		}
		
		return nil, fmt.Errorf("failed to convert to team participation: %w", err)
	}

	return &AcceptTeamInvitationOut{}, nil
}
