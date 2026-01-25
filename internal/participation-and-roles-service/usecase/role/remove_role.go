package role

import (
	"context"
	"fmt"

	rolepolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type RemoveRoleIn struct {
	HackathonID  uuid.UUID
	TargetUserID uuid.UUID
	Role         string
}

type RemoveRoleOut struct{}

type removePolicyRepositoryAdapter struct {
	roleRepo StaffRoleRepository
}

func (a *removePolicyRepositoryAdapter) GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error) {
	return a.roleRepo.GetRoleStrings(ctx, hackathonID, userID)
}

func (a *removePolicyRepositoryAdapter) HasRole(ctx context.Context, hackathonID, userID uuid.UUID, role string) (bool, error) {
	return a.roleRepo.HasRole(ctx, hackathonID, userID, role)
}

func (s *Service) RemoveRole(ctx context.Context, in RemoveRoleIn) (*RemoveRoleOut, error) {
	_, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	adapter := &removePolicyRepositoryAdapter{
		roleRepo: s.roleRepo,
	}

	removePolicy := rolepolicy.NewRemoveHackathonRolePolicy(adapter)
	pctx, err := removePolicy.LoadContext(ctx, rolepolicy.RemoveHackathonRoleParams{
		HackathonID:  in.HackathonID,
		TargetUserID: in.TargetUserID,
		Role:         in.Role,
	})
	if err != nil {
		return nil, err
	}

	decision := removePolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, s.mapPolicyError(decision)
	}

	err = s.roleRepo.Delete(ctx, in.HackathonID, in.TargetUserID, in.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to remove role: %w", err)
	}

	return &RemoveRoleOut{}, nil
}
