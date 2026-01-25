package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type StaffRoleRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewStaffRoleRepository(db queries.DBTX) *StaffRoleRepository {
	return &StaffRoleRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *StaffRoleRepository) Create(ctx context.Context, role *entity.StaffRole) error {
	err := r.Queries().CreateStaffRole(ctx, queries.CreateStaffRoleParams{
		HackathonID: role.HackathonID,
		UserID:      role.UserID,
		Role:        role.Role,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsConflict(err) {
			return nil
		}
		return fmt.Errorf("failed to create staff role: %w", err)
	}
	return nil
}

func (r *StaffRoleRepository) GetByHackathonID(ctx context.Context, hackathonID uuid.UUID) ([]*entity.StaffRole, error) {
	rows, err := r.Queries().GetStaffRolesByHackathonID(ctx, hackathonID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get staff roles by hackathon: %w", err)
	}

	roles := make([]*entity.StaffRole, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, &entity.StaffRole{
			HackathonID: row.HackathonID,
			UserID:      row.UserID,
			Role:        row.Role,
			CreatedAt:   row.CreatedAt,
		})
	}

	return roles, nil
}

func (r *StaffRoleRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.StaffRole, error) {
	rows, err := r.Queries().GetStaffRolesByUserID(ctx, userID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get staff roles by user: %w", err)
	}

	roles := make([]*entity.StaffRole, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, &entity.StaffRole{
			HackathonID: row.HackathonID,
			UserID:      row.UserID,
			Role:        row.Role,
			CreatedAt:   row.CreatedAt,
		})
	}

	return roles, nil
}

func (r *StaffRoleRepository) GetByHackathonAndUser(ctx context.Context, hackathonID, userID uuid.UUID) ([]*entity.StaffRole, error) {
	rows, err := r.Queries().GetStaffRolesByHackathonAndUser(ctx, queries.GetStaffRolesByHackathonAndUserParams{
		HackathonID: hackathonID,
		UserID:      userID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get staff roles: %w", err)
	}

	roles := make([]*entity.StaffRole, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, &entity.StaffRole{
			HackathonID: row.HackathonID,
			UserID:      row.UserID,
			Role:        row.Role,
			CreatedAt:   row.CreatedAt,
		})
	}

	return roles, nil
}

func (r *StaffRoleRepository) Delete(ctx context.Context, hackathonID, userID uuid.UUID, role string) error {
	err := r.Queries().DeleteStaffRole(ctx, queries.DeleteStaffRoleParams{
		HackathonID: hackathonID,
		UserID:      userID,
		Role:        role,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to delete staff role: %w", err)
	}
	return nil
}

func (r *StaffRoleRepository) HasRole(ctx context.Context, hackathonID, userID uuid.UUID, role string) (bool, error) {
	exists, err := r.Queries().HasStaffRole(ctx, queries.HasStaffRoleParams{
		HackathonID: hackathonID,
		UserID:      userID,
		Role:        role,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return false, fmt.Errorf("failed to check staff role: %w", err)
	}
	return exists, nil
}

func (r *StaffRoleRepository) GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error) {
	roles, err := r.GetByHackathonAndUser(ctx, hackathonID, userID)
	if err != nil {
		return nil, err
	}

	roleStrings := make([]string, 0, len(roles))
	for _, role := range roles {
		roleStrings = append(roleStrings, role.Role)
	}

	return roleStrings, nil
}
