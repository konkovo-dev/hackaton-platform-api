package role

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type GetMyHackathonIDsIn struct {
	UserID                    uuid.UUID
	RoleFilter                *string
	ParticipationFilter       *bool
	ParticipationStatusFilter *string
}

type GetMyHackathonIDsOut struct {
	HackathonIDs []uuid.UUID
}

func (s *Service) GetMyHackathonIDs(ctx context.Context, in GetMyHackathonIDsIn) (*GetMyHackathonIDsOut, error) {
	hackathonIDsMap := make(map[uuid.UUID]bool)

	if in.RoleFilter != nil {
		ids, err := s.roleRepo.GetHackathonIDsByUserRole(ctx, in.UserID, *in.RoleFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to get hackathon IDs by role: %w", err)
		}
		for _, id := range ids {
			hackathonIDsMap[id] = true
		}
	} else if in.ParticipationStatusFilter != nil {
		ids, err := s.participRepo.GetHackathonIDsByUserParticipationStatus(ctx, in.UserID, *in.ParticipationStatusFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to get hackathon IDs by participation status: %w", err)
		}
		for _, id := range ids {
			hackathonIDsMap[id] = true
		}
	} else if in.ParticipationFilter != nil && *in.ParticipationFilter {
		ids, err := s.participRepo.GetHackathonIDsByUserParticipation(ctx, in.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get hackathon IDs by participation: %w", err)
		}
		for _, id := range ids {
			hackathonIDsMap[id] = true
		}
	} else {
		roleIDs, err := s.roleRepo.GetHackathonIDsByUserAnyRole(ctx, in.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get hackathon IDs by any role: %w", err)
		}
		for _, id := range roleIDs {
			hackathonIDsMap[id] = true
		}

		participIDs, err := s.participRepo.GetHackathonIDsByUserParticipation(ctx, in.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get hackathon IDs by participation: %w", err)
		}
		for _, id := range participIDs {
			hackathonIDsMap[id] = true
		}
	}

	hackathonIDs := make([]uuid.UUID, 0, len(hackathonIDsMap))
	for id := range hackathonIDsMap {
		hackathonIDs = append(hackathonIDs, id)
	}

	return &GetMyHackathonIDsOut{
		HackathonIDs: hackathonIDs,
	}, nil
}
