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

type AcceptJoinRequestIn struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
	RequestID   uuid.UUID
}

type AcceptJoinRequestOut struct{}

func (s *Service) AcceptJoinRequest(ctx context.Context, in AcceptJoinRequestIn) (*AcceptJoinRequestOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	request, err := s.joinRequestRepo.GetByID(ctx, in.RequestID)
	if err != nil {
		return nil, ErrNotFound
	}

	if request.TeamID != in.TeamID || request.HackathonID != in.HackathonID {
		return nil, ErrNotFound
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

	vacancy, err := s.vacancyRepo.GetByID(ctx, request.VacancyID)
	if err != nil {
		return nil, fmt.Errorf("%w: vacancy not found", ErrNotFound)
	}

	requesterParticipationStatus, err := s.parClient.GetUserParticipation(ctx, in.HackathonID.String(), request.RequesterUserID.String())
	if err != nil {
		s.logger.Error("failed to get requester participation", "error", err)
		return nil, fmt.Errorf("failed to get requester participation: %w", err)
	}

	requesterIsTeamMember := false
	status := strings.ToLower(requesterParticipationStatus)
	if status == "part_team_member" || status == "part_team_captain" {
		requesterIsTeamMember = true
	}

	acceptPolicy := policy.NewAcceptJoinRequestPolicy()
	pctx, err := acceptPolicy.LoadContext(ctx, policy.AcceptJoinRequestParams{
		HackathonID: in.HackathonID,
		TeamID:      in.TeamID,
		RequestID:   in.RequestID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetAllowTeam(allowTeam)
	pctx.SetIsCaptain(isCaptain)
	pctx.SetRequestStatus(request.Status)
	pctx.SetRequesterIsStaff(false)
	pctx.SetRequesterIsTeamMember(requesterIsTeamMember)
	pctx.SetSlotsOpen(vacancy.SlotsOpen)

	decision := acceptPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	now := time.Now().UTC()
	membership := &entity.Membership{
		TeamID:            in.TeamID,
		UserID:            request.RequesterUserID,
		IsCaptain:         false,
		AssignedVacancyID: &request.VacancyID,
		JoinedAt:          now,
	}

	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		joinRequestRepoTx := postgres.NewJoinRequestRepository(tx)
		if err := joinRequestRepoTx.UpdateStatus(ctx, in.RequestID, "accepted"); err != nil {
			return fmt.Errorf("failed to accept join request: %w", err)
		}

		membershipRepoTx := postgres.NewMembershipRepository(tx)
		if err := membershipRepoTx.Create(ctx, membership); err != nil {
			return fmt.Errorf("failed to create membership: %w", err)
		}

		vacancyRepoTx := postgres.NewVacancyRepository(tx)
		if err := vacancyRepoTx.DecrementSlotsOpen(ctx, request.VacancyID); err != nil {
			return fmt.Errorf("failed to decrement slots: %w", err)
		}

		invitationRepoTx := postgres.NewTeamInvitationRepository(tx)
		if err := invitationRepoTx.CancelCompeting(ctx, request.RequesterUserID, in.HackathonID); err != nil {
			return fmt.Errorf("failed to cancel competing invitations: %w", err)
		}

		if err := joinRequestRepoTx.CancelCompeting(ctx, request.RequesterUserID, in.HackathonID); err != nil {
			return fmt.Errorf("failed to cancel competing join requests: %w", err)
		}

		return nil
	})
	if err != nil {
		s.logger.Error("failed to accept join request transaction", "error", err)
		return nil, fmt.Errorf("failed to accept join request: %w", err)
	}

	err = s.parClient.ConvertToTeamParticipation(ctx, in.HackathonID.String(), request.RequesterUserID.String(), in.TeamID.String(), false)
	if err != nil {
		s.logger.Error("failed to convert to team participation, compensating", "error", err, "hackathon_id", in.HackathonID, "team_id", in.TeamID, "requester_id", request.RequesterUserID)

		compErr := s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
			reqRepoTx := postgres.NewJoinRequestRepository(tx)
			if err := reqRepoTx.UpdateStatus(ctx, in.RequestID, "pending"); err != nil {
				return err
			}

			membershipRepoTx := postgres.NewMembershipRepository(tx)
			if err := membershipRepoTx.Delete(ctx, in.TeamID, request.RequesterUserID); err != nil {
				return err
			}

			vacancyRepoTx := postgres.NewVacancyRepository(tx)
			if err := vacancyRepoTx.IncrementSlotsOpen(ctx, request.VacancyID); err != nil {
				return err
			}

			return nil
		})

		if compErr != nil {
			s.logger.Error("CRITICAL: failed to compensate join request acceptance, data inconsistency", "error", compErr, "hackathon_id", in.HackathonID, "team_id", in.TeamID, "requester_id", request.RequesterUserID, "request_id", in.RequestID)
		}

		return nil, fmt.Errorf("failed to convert to team participation: %w", err)
	}

	return &AcceptJoinRequestOut{}, nil
}
