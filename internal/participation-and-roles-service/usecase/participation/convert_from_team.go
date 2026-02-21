package participation

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/google/uuid"
)

type ConvertFromTeamIn struct {
	HackathonID uuid.UUID
	UserID      uuid.UUID
}

type ConvertFromTeamOut struct {
	NewStatus string
}

func (s *Service) ConvertFromTeam(ctx context.Context, in ConvertFromTeamIn) (*ConvertFromTeamOut, error) {
	participation, err := s.participRepo.Get(ctx, in.HackathonID, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participation: %w", err)
	}

	if participation == nil {
		return nil, ErrNotFound
	}

	if participation.Status != string(domain.ParticipationTeamMember) &&
		participation.Status != string(domain.ParticipationTeamCaptain) {
		return nil, fmt.Errorf("%w: can only convert from TEAM_MEMBER or TEAM_CAPTAIN", ErrInvalidInput)
	}

	newStatus := string(domain.ParticipationLookingForTeam)

	err = s.participRepo.Update(ctx, in.HackathonID, in.UserID, newStatus, nil, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to update participation: %w", err)
	}

	return &ConvertFromTeamOut{
		NewStatus: newStatus,
	}, nil
}
