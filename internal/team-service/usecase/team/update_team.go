package team

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	teampolicy "github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	outboxusecase "github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/google/uuid"
)

type UpdateTeamIn struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
	Name        string
	Description string
	IsJoinable  bool
}

type UpdateTeamOut struct {
	Team *entity.Team
}

func (s *Service) UpdateTeam(ctx context.Context, in UpdateTeamIn) (*UpdateTeamOut, error) {
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

	oldTeam, err := s.teamRepo.GetByIDAndHackathonID(ctx, in.TeamID, in.HackathonID)
	if err != nil {
		return nil, ErrNotFound
	}

	isCaptain, err := s.membershipRepo.CheckIsCaptain(ctx, in.TeamID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to check captain status: %w", err)
	}

	updatePolicy := teampolicy.NewUpdateTeamPolicy()
	pctx, err := updatePolicy.LoadContext(ctx, teampolicy.UpdateTeamParams{
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

	decision := updatePolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	if in.Name != oldTeam.Name {
		exists, err := s.teamRepo.CheckNameExists(ctx, in.HackathonID, in.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check team name: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("%w: team name already exists in this hackathon", ErrConflict)
		}
	}

	updatedTeam := &entity.Team{
		ID:          oldTeam.ID,
		HackathonID: oldTeam.HackathonID,
		Name:        in.Name,
		Description: in.Description,
		IsJoinable:  in.IsJoinable,
		CreatedAt:   oldTeam.CreatedAt,
	}

	err = s.teamRepo.Update(ctx, updatedTeam)
	if err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	eventPayload := outboxusecase.TeamUpdatedPayload{
		TeamID:      updatedTeam.ID.String(),
		HackathonID: updatedTeam.HackathonID.String(),
		Name:        updatedTeam.Name,
		Description: updatedTeam.Description,
		IsJoinable:  updatedTeam.IsJoinable,
		UpdatedAt:   updatedTeam.UpdatedAt,
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event payload: %w", err)
	}

	outboxEvent := &outbox.Event{
		ID:            uuid.New(),
		AggregateID:   updatedTeam.ID.String(),
		AggregateType: "team",
		EventType:     outboxusecase.EventTypeTeamUpdated,
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

	return &UpdateTeamOut{
		Team: updatedTeam,
	}, nil
}
