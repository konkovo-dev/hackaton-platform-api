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

type CreateJoinRequestIn struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
	VacancyID   uuid.UUID
	Message     string
}

type CreateJoinRequestOut struct {
	RequestID uuid.UUID
}

func (s *Service) CreateJoinRequest(ctx context.Context, in CreateJoinRequestIn) (*CreateJoinRequestOut, error) {
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

	vacancy, err := s.vacancyRepo.GetByID(ctx, in.VacancyID)
	if err != nil {
		return nil, fmt.Errorf("%w: vacancy not found", ErrNotFound)
	}

	if vacancy.TeamID != in.TeamID {
		return nil, fmt.Errorf("%w: vacancy does not belong to team", ErrBadRequest)
	}

	_, participationStatus, roles, err := s.parClient.GetHackathonContext(ctx, in.HackathonID.String())
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

	createPolicy := policy.NewCreateJoinRequestPolicy()
	pctx, err := createPolicy.LoadContext(ctx, policy.CreateJoinRequestParams{
		HackathonID: in.HackathonID,
		TeamID:      in.TeamID,
		VacancyID:   in.VacancyID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetAllowTeam(allowTeam)
	pctx.SetTeamIsJoinable(team.IsJoinable)
	pctx.SetIsStaff(isStaff)
	pctx.SetParticipationStatus(participationStatus)
	pctx.SetSlotsOpen(vacancy.SlotsOpen)

	decision := createPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	now := time.Now().UTC()
	requestID := uuid.New()

	request := &entity.JoinRequest{
		ID:              requestID,
		HackathonID:     in.HackathonID,
		TeamID:          in.TeamID,
		VacancyID:       in.VacancyID,
		RequesterUserID: userUUID,
		Message:         in.Message,
		Status:          "pending",
		CreatedAt:       now,
		UpdatedAt:       now,
		ExpiresAt:       nil,
	}

	err = s.joinRequestRepo.Create(ctx, request)
	if err != nil {
		s.logger.Error("failed to create join request", "error", err)
		return nil, fmt.Errorf("failed to create join request: %w", err)
	}

	return &CreateJoinRequestOut{
		RequestID: requestID,
	}, nil
}
