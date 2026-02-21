package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/participation"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

func (r *ParticipationRepository) List(ctx context.Context, hackathonID uuid.UUID, filter participation.ListParticipationsFilter) ([]*entity.Participation, int64, error) {
	var rows []queries.ParticipationAndRolesParticipation
	var total int64
	var err error

	if len(filter.Statuses) > 0 {
		rows, err = r.Queries().ListParticipationsByStatus(ctx, queries.ListParticipationsByStatusParams{
			HackathonID: hackathonID,
			Column2:     filter.Statuses,
			Limit:       filter.Limit,
			Offset:      filter.Offset,
		})
		if err != nil {
			err = pgxutil.MapDBError(err)
			return nil, 0, fmt.Errorf("failed to list participations by status: %w", err)
		}

		total, err = r.Queries().CountParticipationsByStatus(ctx, queries.CountParticipationsByStatusParams{
			HackathonID: hackathonID,
			Column2:     filter.Statuses,
		})
		if err != nil {
			err = pgxutil.MapDBError(err)
			return nil, 0, fmt.Errorf("failed to count participations by status: %w", err)
		}
	} else {
		rows, err = r.Queries().ListParticipations(ctx, queries.ListParticipationsParams{
			HackathonID: hackathonID,
			Limit:       filter.Limit,
			Offset:      filter.Offset,
		})
		if err != nil {
			err = pgxutil.MapDBError(err)
			return nil, 0, fmt.Errorf("failed to list participations: %w", err)
		}

		total, err = r.Queries().CountParticipations(ctx, hackathonID)
		if err != nil {
			err = pgxutil.MapDBError(err)
			return nil, 0, fmt.Errorf("failed to count participations: %w", err)
		}
	}

	participations := make([]*entity.Participation, 0, len(rows))
	for _, row := range rows {
		var teamID *uuid.UUID
		if row.TeamID.Valid {
			id := uuid.UUID(row.TeamID.Bytes)
			teamID = &id
		}

		participations = append(participations, &entity.Participation{
			HackathonID:    row.HackathonID,
			UserID:         row.UserID,
			Status:         row.Status,
			TeamID:         teamID,
			WishedRoles:    nil,
			MotivationText: row.MotivationText,
			RegisteredAt:   row.RegisteredAt,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
		})
	}

	if len(filter.WishedRoleIDs) > 0 {
		participations = r.filterByWishedRoles(ctx, participations, filter.WishedRoleIDs)
		total = int64(len(participations))
	}

	return participations, total, nil
}

func (r *ParticipationRepository) filterByWishedRoles(ctx context.Context, participations []*entity.Participation, roleIDs []uuid.UUID) []*entity.Participation {
	filtered := make([]*entity.Participation, 0)

	for _, p := range participations {
		roles, err := r.Queries().GetWishedRolesByParticipation(ctx, queries.GetWishedRolesByParticipationParams{
			HackathonID: p.HackathonID,
			UserID:      p.UserID,
		})
		if err != nil {
			continue
		}

		hasRole := false
		for _, role := range roles {
			for _, filterRoleID := range roleIDs {
				if role.ID == filterRoleID {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if hasRole {
			filtered = append(filtered, p)
		}
	}

	return filtered
}
