package team

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	outboxusecase "github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/google/uuid"
)

type DeleteTeamIn struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
}

type DeleteTeamOut struct{}

func (s *Service) DeleteTeam(ctx context.Context, in DeleteTeamIn) (*DeleteTeamOut, error) {
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

	membersCount, err := s.membershipRepo.CountMembers(ctx, in.TeamID)
	if err != nil {
		s.logger.Error("failed to count members", "error", err)
		return nil, fmt.Errorf("failed to count members: %w", err)
	}

	deletePolicy := policy.NewDeleteTeamPolicy()
	pctx, err := deletePolicy.LoadContext(ctx, policy.DeleteTeamParams{
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
	pctx.SetMembersCount(membersCount)

	decision := deletePolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	err = s.parClient.ConvertFromTeamParticipation(ctx, in.HackathonID.String(), userUUID.String())
	if err != nil {
		s.logger.Error("failed to convert from team participation", "error", err)
		return nil, fmt.Errorf("failed to convert from team participation: %w", err)
	}

	err = s.teamRepo.Delete(ctx, team.ID)
	if err != nil {
		s.logger.Error("failed to delete team", "error", err)
		return nil, fmt.Errorf("failed to delete team: %w", err)
	}

	eventPayload := outboxusecase.TeamDeletedPayload{
		TeamID:      team.ID.String(),
		HackathonID: team.HackathonID.String(),
		DeletedAt:   time.Now().UTC(),
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event payload: %w", err)
	}

	outboxEvent := &outbox.Event{
		ID:            uuid.New(),
		AggregateID:   team.ID.String(),
		AggregateType: "team",
		EventType:     outboxusecase.EventTypeTeamDeleted,
		Payload:       payloadBytes,
		Status:        outbox.EventStatusPending,
		AttemptCount:  0,
		LastError:     "",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := s.outboxRepo.Create(ctx, outboxEvent); err != nil {
		return nil, fmt.Errorf("failed to create outbox event: %w", err)
	}

	return &DeleteTeamOut{}, nil
}
