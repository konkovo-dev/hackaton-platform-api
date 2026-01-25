package role

import (
	"context"
	"fmt"

	rolepolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type SelfRemoveRoleIn struct {
	HackathonID uuid.UUID
	Role        string
}

type SelfRemoveRoleOut struct{}

type selfRemovePolicyRepositoryAdapter struct {
	roleRepo StaffRoleRepository
}

func (a *selfRemovePolicyRepositoryAdapter) HasRole(ctx context.Context, hackathonID, userID uuid.UUID, role string) (bool, error) {
	return a.roleRepo.HasRole(ctx, hackathonID, userID, role)
}

func (s *Service) SelfRemoveRole(ctx context.Context, in SelfRemoveRoleIn) (*SelfRemoveRoleOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	adapter := &selfRemovePolicyRepositoryAdapter{
		roleRepo: s.roleRepo,
	}

	selfRemovePolicy := rolepolicy.NewSelfRemoveHackathonRolePolicy(adapter)
	pctx, err := selfRemovePolicy.LoadContext(ctx, rolepolicy.SelfRemoveHackathonRoleParams{
		HackathonID: in.HackathonID,
		Role:        in.Role,
	})
	if err != nil {
		return nil, err
	}

	decision := selfRemovePolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, s.mapPolicyError(decision)
	}

	err = s.roleRepo.Delete(ctx, in.HackathonID, userUUID, in.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to self remove role: %w", err)
	}

	return &SelfRemoveRoleOut{}, nil
}
