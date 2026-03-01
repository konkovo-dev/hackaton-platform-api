package participation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	participationpolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	outboxusecase "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/google/uuid"
)

type UpdateMyIn struct {
	HackathonID    uuid.UUID
	WishedRoleIDs  []uuid.UUID
	MotivationText string
}

type UpdateMyOut struct {
	Participation *entity.Participation
}

func (s *Service) UpdateMy(ctx context.Context, in UpdateMyIn) (*UpdateMyOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	adapter := &policyRepositoryAdapter{
		roleRepo:     s.roleRepo,
		participRepo: s.participRepo,
	}

	updatePolicy := participationpolicy.NewUpdateMyParticipationPolicy(adapter)
	pctx, err := updatePolicy.LoadContext(ctx, participationpolicy.UpdateMyParticipationParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := updatePolicy.Check(ctx, pctx)
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

	var wishedRoles []*entity.TeamRole
	if len(in.WishedRoleIDs) > 0 {
		wishedRoles, err = s.teamRoleRepo.GetByIDs(ctx, in.WishedRoleIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get team roles: %w", err)
		}
		if len(wishedRoles) != len(in.WishedRoleIDs) {
			return nil, fmt.Errorf("%w: some team role IDs are invalid", ErrInvalidInput)
		}
	}

	err = s.participRepo.UpdateProfile(ctx, in.HackathonID, userUUID, in.MotivationText, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to update participation profile: %w", err)
	}

	err = s.teamRoleRepo.SetForParticipation(ctx, in.HackathonID, userUUID, in.WishedRoleIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to set wished roles: %w", err)
	}

	updated, err := s.participRepo.Get(ctx, in.HackathonID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated participation: %w", err)
	}

	updated.WishedRoles = wishedRoles

	wishedRoleIDStrs := make([]string, len(in.WishedRoleIDs))
	for i, roleID := range in.WishedRoleIDs {
		wishedRoleIDStrs[i] = roleID.String()
	}

	eventPayload := outboxusecase.ParticipationUpdatedPayload{
		HackathonID:    in.HackathonID.String(),
		UserID:         userUUID.String(),
		WishedRoleIDs:  wishedRoleIDStrs,
		MotivationText: in.MotivationText,
		UpdatedAt:      updated.UpdatedAt,
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event payload: %w", err)
	}

	outboxEvent := &outbox.Event{
		ID:            uuid.New(),
		AggregateID:   fmt.Sprintf("%s:%s", in.HackathonID.String(), userUUID.String()),
		AggregateType: "participation",
		EventType:     outboxusecase.EventTypeParticipationUpdated,
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

	return &UpdateMyOut{
		Participation: updated,
	}, nil
}
