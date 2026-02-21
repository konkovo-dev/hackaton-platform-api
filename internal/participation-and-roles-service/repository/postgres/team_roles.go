package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type TeamRoleRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewTeamRoleRepository(db queries.DBTX) *TeamRoleRepository {
	return &TeamRoleRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *TeamRoleRepository) ListAll(ctx context.Context) ([]*entity.TeamRole, error) {
	rows, err := r.Queries().ListAllTeamRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list team roles: %w", err)
	}

	roles := make([]*entity.TeamRole, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, &entity.TeamRole{
			ID:   row.ID,
			Name: row.Name,
		})
	}

	return roles, nil
}

func (r *TeamRoleRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.TeamRole, error) {
	if len(ids) == 0 {
		return []*entity.TeamRole{}, nil
	}

	rows, err := r.Queries().GetTeamRolesByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get team roles: %w", err)
	}

	roles := make([]*entity.TeamRole, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, &entity.TeamRole{
			ID:   row.ID,
			Name: row.Name,
		})
	}

	return roles, nil
}

func (r *TeamRoleRepository) GetByParticipation(ctx context.Context, hackathonID, userID uuid.UUID) ([]*entity.TeamRole, error) {
	rows, err := r.Queries().GetWishedRolesByParticipation(ctx, queries.GetWishedRolesByParticipationParams{
		HackathonID: hackathonID,
		UserID:      userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get wished roles: %w", err)
	}

	roles := make([]*entity.TeamRole, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, &entity.TeamRole{
			ID:   row.ID,
			Name: row.Name,
		})
	}

	return roles, nil
}

func (r *TeamRoleRepository) SetForParticipation(ctx context.Context, hackathonID, userID uuid.UUID, roleIDs []uuid.UUID) error {
	err := r.Queries().DeleteWishedRolesByParticipation(ctx, queries.DeleteWishedRolesByParticipationParams{
		HackathonID: hackathonID,
		UserID:      userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete wished roles: %w", err)
	}

	for _, roleID := range roleIDs {
		err := r.Queries().CreateWishedRole(ctx, queries.CreateWishedRoleParams{
			HackathonID: hackathonID,
			UserID:      userID,
			TeamRoleID:  roleID,
		})
		if err != nil {
			err = pgxutil.MapDBError(err)
			if pgxutil.IsNotFound(err) {
				return fmt.Errorf("team role not found: %s: %w", roleID, err)
			}
			return fmt.Errorf("failed to create wished role: %w", err)
		}
	}

	return nil
}
