package role

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	"github.com/google/uuid"
)

type AssignRoleIn struct {
	HackathonID uuid.UUID
	UserID      uuid.UUID
	Role        string
}

func (s *Service) AssignRole(ctx context.Context, in AssignRoleIn) error {
	role := &entity.StaffRole{
		HackathonID: in.HackathonID,
		UserID:      in.UserID,
		Role:        in.Role,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return fmt.Errorf("failed to create staff role: %w", err)
	}

	return nil
}
