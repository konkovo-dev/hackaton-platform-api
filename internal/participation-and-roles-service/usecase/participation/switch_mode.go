package participation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	participationpolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	outboxusecase "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/google/uuid"
)

type SwitchModeIn struct {
	HackathonID uuid.UUID
	NewStatus   string
}

type SwitchModeOut struct {
	Participation *entity.Participation
}

func (s *Service) SwitchMode(ctx context.Context, in SwitchModeIn) (*SwitchModeOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if in.NewStatus != string(domain.ParticipationIndividual) &&
		in.NewStatus != string(domain.ParticipationLookingForTeam) {
		return nil, fmt.Errorf("%w: new_status must be INDIVIDUAL_ACTIVE or LOOKING_FOR_TEAM", ErrInvalidInput)
	}

	adapter := &policyRepositoryAdapter{
		roleRepo:     s.roleRepo,
		participRepo: s.participRepo,
	}

	switchPolicy := participationpolicy.NewSwitchParticipationModePolicy(adapter)
	pctx, err := switchPolicy.LoadContext(ctx, participationpolicy.SwitchParticipationModeParams{
		HackathonID: in.HackathonID,
		NewStatus:   in.NewStatus,
	})
	if err != nil {
		return nil, err
	}

	decision := switchPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	existing, err := s.participRepo.Get(ctx, in.HackathonID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participation: %w", err)
	}

	if existing == nil {
		return nil, ErrNotFound
	}

	err = s.participRepo.UpdateStatus(ctx, in.HackathonID, userUUID, in.NewStatus, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to update participation status: %w", err)
	}

	updated, err := s.participRepo.Get(ctx, in.HackathonID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated participation: %w", err)
	}

	wishedRoles, err := s.teamRoleRepo.GetByParticipation(ctx, in.HackathonID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wished roles: %w", err)
	}

	updated.WishedRoles = wishedRoles

	eventPayload := outboxusecase.ParticipationStatusChangedPayload{
		HackathonID: in.HackathonID.String(),
		UserID:      userUUID.String(),
		OldStatus:   existing.Status,
		NewStatus:   in.NewStatus,
		UpdatedAt:   updated.UpdatedAt,
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event payload: %w", err)
	}

	outboxEvent := &outbox.Event{
		ID:            uuid.New(),
		AggregateID:   fmt.Sprintf("%s:%s", in.HackathonID.String(), userUUID.String()),
		AggregateType: "participation",
		EventType:     outboxusecase.EventTypeParticipationStatusChanged,
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

	return &SwitchModeOut{
		Participation: updated,
	}, nil
}
