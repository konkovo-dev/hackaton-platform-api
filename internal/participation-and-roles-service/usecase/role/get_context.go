package role

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type GetHackathonContextIn struct {
	UserID      uuid.UUID
	HackathonID uuid.UUID
}

type GetHackathonContextOut struct {
	UserID              uuid.UUID
	Roles               []string
	ParticipationStatus string
	TeamID              string
}

func (s *Service) GetHackathonContext(ctx context.Context, in GetHackathonContextIn) (*GetHackathonContextOut, error) {
	roles, err := s.roleRepo.GetByHackathonAndUser(ctx, in.HackathonID, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get staff roles: %w", err)
	}

	roleStrings := make([]string, 0, len(roles))
	for _, role := range roles {
		roleStrings = append(roleStrings, role.Role)
	}

	return &GetHackathonContextOut{
		UserID:              in.UserID,
		Roles:               roleStrings,
		ParticipationStatus: "none",
		TeamID:              "",
	}, nil
}
