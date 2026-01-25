package role

import (
	"context"
	"fmt"

	rolepolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/google/uuid"
)

type ListStaffIn struct {
	HackathonID uuid.UUID
}

type StaffMember struct {
	UserID uuid.UUID
	Roles  []string
}

type ListStaffOut struct {
	Staff []*StaffMember
}

func (s *Service) ListStaff(ctx context.Context, in ListStaffIn) (*ListStaffOut, error) {
	listPolicy := rolepolicy.NewListStaffPolicy(s.roleRepo)
	pctx, err := listPolicy.LoadContext(ctx, rolepolicy.ListStaffParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	decision := listPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, s.mapPolicyError(decision)
	}

	roles, err := s.roleRepo.GetByHackathonID(ctx, in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get staff roles: %w", err)
	}

	staffMap := make(map[uuid.UUID][]string)
	for _, role := range roles {
		staffMap[role.UserID] = append(staffMap[role.UserID], role.Role)
	}

	staff := make([]*StaffMember, 0, len(staffMap))
	for userID, userRoles := range staffMap {
		staff = append(staff, &StaffMember{
			UserID: userID,
			Roles:  userRoles,
		})
	}

	return &ListStaffOut{
		Staff: staff,
	}, nil
}
