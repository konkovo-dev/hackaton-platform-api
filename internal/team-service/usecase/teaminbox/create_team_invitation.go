package teaminbox

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type CreateTeamInvitationIn struct {
	HackathonID  uuid.UUID
	TeamID       uuid.UUID
	TargetUserID uuid.UUID
	VacancyID    uuid.UUID
	Message      string
}

type CreateTeamInvitationOut struct {
	InvitationID uuid.UUID
}

func (s *Service) CreateTeamInvitation(ctx context.Context, in CreateTeamInvitationIn) (*CreateTeamInvitationOut, error) {
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

	team, err := s.teamRepo.GetByIDAndHackathonID(ctx, in.TeamID, in.HackathonID)
	if err != nil {
		return nil, ErrNotFound
	}

	isCaptain, err := s.membershipRepo.CheckIsCaptain(ctx, in.TeamID, userUUID)
	if err != nil {
		s.logger.Error("failed to check captain status", "error", err)
		return nil, fmt.Errorf("failed to check captain status: %w", err)
	}

	vacancy, err := s.vacancyRepo.GetByID(ctx, in.VacancyID)
	if err != nil {
		return nil, fmt.Errorf("%w: vacancy not found", ErrBadRequest)
	}

	if vacancy.TeamID != in.TeamID {
		return nil, fmt.Errorf("%w: vacancy does not belong to this team", ErrBadRequest)
	}

	if vacancy.SlotsOpen <= 0 {
		return nil, fmt.Errorf("%w: no open slots in vacancy", ErrConflict)
	}

	targetParticipationStatus, err := s.parClient.GetUserParticipation(ctx, in.HackathonID.String(), in.TargetUserID.String())
	if err != nil {
		s.logger.Error("failed to get target user participation", "error", err)
		return nil, fmt.Errorf("failed to get target user participation: %w", err)
	}

	targetIsTeamMember := false
	status := strings.ToLower(targetParticipationStatus)
	if status == "team_member" || status == "team_captain" {
		targetIsTeamMember = true
	}

	createPolicy := policy.NewCreateTeamInvitationPolicy()
	pctx, err := createPolicy.LoadContext(ctx, policy.CreateTeamInvitationParams{
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
	pctx.SetIsCaptain(isCaptain)
	pctx.SetTargetIsStaff(false)
	pctx.SetIsTargetTeamMember(targetIsTeamMember)

	decision := createPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	invitationID := uuid.New()
	now := time.Now().UTC()

	invitation := &entity.TeamInvitation{
		ID:              invitationID,
		HackathonID:     in.HackathonID,
		TeamID:          team.ID,
		VacancyID:       in.VacancyID,
		TargetUserID:    in.TargetUserID,
		CreatedByUserID: userUUID,
		Message:         in.Message,
		Status:          "pending",
		CreatedAt:       now,
		UpdatedAt:       now,
		ExpiresAt:       nil,
	}

	err = s.invitationRepo.Create(ctx, invitation)
	if err != nil {
		s.logger.Error("failed to create team invitation", "error", err)
		return nil, fmt.Errorf("failed to create team invitation: %w", err)
	}

	return &CreateTeamInvitationOut{InvitationID: invitationID}, nil
}
