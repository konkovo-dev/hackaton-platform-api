package me

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	identitypolicy "github.com/belikoooova/hackaton-platform-api/internal/identity-service/policy"
	outboxusecase "github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/google/uuid"
)

type UpdateMySkillsIn struct {
	UserID           uuid.UUID
	CatalogSkillIDs  []uuid.UUID
	CustomSkills     []string
	SkillsVisibility domain.VisibilityLevel
}

type UpdateMySkillsOut struct {
	CatalogSkills []*entity.CatalogSkill
	CustomSkills  []*entity.CustomSkill
	Visibility    *entity.Visibility
}

func (s *Service) UpdateMySkills(ctx context.Context, in UpdateMySkillsIn) (*UpdateMySkillsOut, error) {
	updatePolicy := identitypolicy.NewUpdateMySkillsPolicy()
	pctx, err := updatePolicy.LoadContext(ctx, identitypolicy.UpdateMySkillsParams{
		UserID: in.UserID,
	})
	if err != nil {
		return nil, err
	}

	decision := updatePolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, s.mapPolicyError(decision)
	}

	if err := s.validateUpdateMySkillsIn(in); err != nil {
		return nil, err
	}

	_, err = s.userRepo.GetByID(ctx, in.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	var catalogSkills []*entity.CatalogSkill
	var customSkills []*entity.CustomSkill
	var visibility *entity.Visibility

	err = s.uow.Do(ctx, func(ctx context.Context, txRepos *TxRepositories) error {
		if err := txRepos.Skills.Update(ctx, in.UserID, in.CatalogSkillIDs, in.CustomSkills); err != nil {
			return fmt.Errorf("failed to update user skills: %w", err)
		}

		catalogSkills, err = txRepos.Skills.GetUserCatalogSkills(ctx, in.UserID)
		if err != nil {
			return fmt.Errorf("failed to get catalog skills: %w", err)
		}

		customSkills, err = txRepos.Skills.GetUserCustomSkills(ctx, in.UserID)
		if err != nil {
			return fmt.Errorf("failed to get custom skills: %w", err)
		}

		visibility, err = txRepos.Visibility.GetByUserID(ctx, in.UserID)
		if err != nil {
			return fmt.Errorf("failed to get visibility: %w", err)
		}

		visibility.SkillsVisibility = in.SkillsVisibility

		if err := txRepos.Visibility.Upsert(ctx, visibility); err != nil {
			return fmt.Errorf("failed to update visibility: %w", err)
		}

		catalogSkillIDStrs := make([]string, len(in.CatalogSkillIDs))
		for i, skillID := range in.CatalogSkillIDs {
			catalogSkillIDStrs[i] = skillID.String()
		}

		eventPayload := outboxusecase.UserSkillsUpdatedPayload{
			UserID:           in.UserID.String(),
			CatalogSkillIDs:  catalogSkillIDStrs,
			CustomSkillNames: in.CustomSkills,
			UpdatedAt:        time.Now().UTC(),
		}

		payloadBytes, err := json.Marshal(eventPayload)
		if err != nil {
			return fmt.Errorf("failed to marshal event payload: %w", err)
		}

		outboxEvent := &outbox.Event{
			ID:            uuid.New(),
			AggregateID:   in.UserID.String(),
			AggregateType: "user",
			EventType:     outboxusecase.EventTypeUserSkillsUpdated,
			Payload:       payloadBytes,
			Status:        outbox.EventStatusPending,
			AttemptCount:  0,
			LastError:     "",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}

		if err := txRepos.Outbox.Create(ctx, outboxEvent); err != nil {
			return fmt.Errorf("failed to create outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &UpdateMySkillsOut{
		CatalogSkills: catalogSkills,
		CustomSkills:  customSkills,
		Visibility:    visibility,
	}, nil
}

func (s *Service) validateUpdateMySkillsIn(in UpdateMySkillsIn) error {
	if in.UserID == uuid.Nil {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}

	if in.SkillsVisibility == "" {
		return fmt.Errorf("%w: skills_visibility is required", ErrInvalidInput)
	}

	if in.SkillsVisibility != domain.VisibilityLevelPublic && in.SkillsVisibility != domain.VisibilityLevelPrivate {
		return fmt.Errorf("%w: skills_visibility must be 'public' or 'private'", ErrInvalidInput)
	}

	return nil
}
