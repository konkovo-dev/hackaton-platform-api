package participation

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	participationpolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
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

	return &UpdateMyOut{
		Participation: updated,
	}, nil
}
