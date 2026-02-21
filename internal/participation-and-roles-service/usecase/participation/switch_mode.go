package participation

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	participationpolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
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

	return &SwitchModeOut{
		Participation: updated,
	}, nil
}
