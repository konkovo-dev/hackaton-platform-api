package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ParticipationRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewParticipationRepository(db queries.DBTX) *ParticipationRepository {
	return &ParticipationRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *ParticipationRepository) Get(ctx context.Context, hackathonID, userID uuid.UUID) (*entity.Participation, error) {
	row, err := r.Queries().GetParticipation(ctx, queries.GetParticipationParams{
		HackathonID: hackathonID,
		UserID:      userID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get participation: %w", err)
	}

	var teamID *uuid.UUID
	if row.TeamID.Valid {
		id := uuid.UUID(row.TeamID.Bytes)
		teamID = &id
	}

	return &entity.Participation{
		HackathonID: row.HackathonID,
		UserID:      row.UserID,
		Status:      row.Status,
		TeamID:      teamID,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func (r *ParticipationRepository) Create(ctx context.Context, participation *entity.Participation) error {
	var teamID pgtype.UUID
	if participation.TeamID != nil {
		teamID = pgtype.UUID{
			Bytes: *participation.TeamID,
			Valid: true,
		}
	}

	err := r.Queries().CreateParticipation(ctx, queries.CreateParticipationParams{
		HackathonID: participation.HackathonID,
		UserID:      participation.UserID,
		Status:      participation.Status,
		TeamID:      teamID,
		CreatedAt:   participation.CreatedAt,
		UpdatedAt:   participation.UpdatedAt,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to create participation: %w", err)
	}
	return nil
}

func (r *ParticipationRepository) GetStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error) {
	participation, err := r.Get(ctx, hackathonID, userID)
	if err != nil {
		return "", err
	}
	if participation == nil {
		return "", nil
	}
	return participation.Status, nil
}
