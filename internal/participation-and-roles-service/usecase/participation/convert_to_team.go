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

type ConvertToTeamIn struct {
	HackathonID uuid.UUID
	UserID      uuid.UUID
	TeamID      uuid.UUID
	IsCaptain   bool
}

type ConvertToTeamOut struct {
	NewStatus string
}

func (s *Service) ConvertToTeam(ctx context.Context, in ConvertToTeamIn) (*ConvertToTeamOut, error) {
	participation, err := s.participRepo.Get(ctx, in.HackathonID, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participation: %w", err)
	}

	if participation == nil {
		return nil, ErrNotFound
	}

	isAlreadyInTeam := participation.Status == string(domain.ParticipationTeamMember) ||
		participation.Status == string(domain.ParticipationTeamCaptain)

	if !isAlreadyInTeam {
		if participation.Status != string(domain.ParticipationIndividual) &&
			participation.Status != string(domain.ParticipationLookingForTeam) {
			return nil, fmt.Errorf("%w: can only convert from INDIVIDUAL_ACTIVE or LOOKING_FOR_TEAM", ErrInvalidInput)
		}
	}

	var newStatus string
	if in.IsCaptain {
		newStatus = string(domain.ParticipationTeamCaptain)
	} else {
		newStatus = string(domain.ParticipationTeamMember)
	}

	now := time.Now().UTC()
	err = s.participRepo.Update(ctx, in.HackathonID, in.UserID, newStatus, &in.TeamID, now)
	if err != nil {
		return nil, fmt.Errorf("failed to update participation: %w", err)
	}

	eventPayload := outboxusecase.ParticipationTeamAssignedPayload{
		HackathonID: in.HackathonID.String(),
		UserID:      in.UserID.String(),
		TeamID:      in.TeamID.String(),
		IsCaptain:   in.IsCaptain,
		AssignedAt:  now,
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event payload: %w", err)
	}

	outboxEvent := &outbox.Event{
		ID:            uuid.New(),
		AggregateID:   fmt.Sprintf("%s:%s", in.HackathonID.String(), in.UserID.String()),
		AggregateType: "participation",
		EventType:     outboxusecase.EventTypeParticipationTeamAssigned,
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

	return &ConvertToTeamOut{
		NewStatus: newStatus,
	}, nil
}
