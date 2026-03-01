package participation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	outboxusecase "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/google/uuid"
)

type ConvertFromTeamIn struct {
	HackathonID uuid.UUID
	UserID      uuid.UUID
}

type ConvertFromTeamOut struct {
	NewStatus string
}

func (s *Service) ConvertFromTeam(ctx context.Context, in ConvertFromTeamIn) (*ConvertFromTeamOut, error) {
	participation, err := s.participRepo.Get(ctx, in.HackathonID, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participation: %w", err)
	}

	if participation == nil {
		return nil, ErrNotFound
	}

	if participation.Status != string(domain.ParticipationTeamMember) &&
		participation.Status != string(domain.ParticipationTeamCaptain) {
		return nil, fmt.Errorf("%w: can only convert from TEAM_MEMBER or TEAM_CAPTAIN", ErrInvalidInput)
	}

	newStatus := string(domain.ParticipationLookingForTeam)
	now := time.Now().UTC()

	oldTeamID := ""
	if participation.TeamID != nil {
		oldTeamID = participation.TeamID.String()
	}

	err = s.participRepo.Update(ctx, in.HackathonID, in.UserID, newStatus, nil, now)
	if err != nil {
		return nil, fmt.Errorf("failed to update participation: %w", err)
	}

	if oldTeamID != "" {
		eventPayload := outboxusecase.ParticipationTeamRemovedPayload{
			HackathonID: in.HackathonID.String(),
			UserID:      in.UserID.String(),
			TeamID:      oldTeamID,
			RemovedAt:   now,
		}

		payloadBytes, err := json.Marshal(eventPayload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal event payload: %w", err)
		}

		outboxEvent := &outbox.Event{
			ID:            uuid.New(),
			AggregateID:   fmt.Sprintf("%s:%s", in.HackathonID.String(), in.UserID.String()),
			AggregateType: "participation",
			EventType:     outboxusecase.EventTypeParticipationTeamRemoved,
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
	}

	return &ConvertFromTeamOut{
		NewStatus: newStatus,
	}, nil
}
