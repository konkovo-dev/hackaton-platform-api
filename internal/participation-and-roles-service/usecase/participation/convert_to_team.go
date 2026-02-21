package participation

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/google/uuid"
)

type ConvertToTeamIn struct {
	HackathonID uuid.UUID
	UserID      uuid.UUID
	TeamID      uuid.UUID
	IsCaptain   bool
}

type ConvertToTeamOut struct {
	NewStatus string
}

func (s *Service) ConvertToTeam(ctx context.Context, in ConvertToTeamIn) (*ConvertToTeamOut, error) {
	participation, err := s.participRepo.Get(ctx, in.HackathonID, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participation: %w", err)
	}

	if participation == nil {
		return nil, ErrNotFound
	}

	if participation.Status != string(domain.ParticipationIndividual) &&
		participation.Status != string(domain.ParticipationLookingForTeam) {
		return nil, fmt.Errorf("%w: can only convert from INDIVIDUAL_ACTIVE or LOOKING_FOR_TEAM", ErrInvalidInput)
	}

	var newStatus string
	if in.IsCaptain {
		newStatus = string(domain.ParticipationTeamCaptain)
	} else {
		newStatus = string(domain.ParticipationTeamMember)
	}

	err = s.participRepo.Update(ctx, in.HackathonID, in.UserID, newStatus, &in.TeamID, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to update participation: %w", err)
	}

	return &ConvertToTeamOut{
		NewStatus: newStatus,
	}, nil
}
