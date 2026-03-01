package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/repository/postgres/queries"
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

func (r *ParticipationRepository) Upsert(ctx context.Context, participation *entity.Participation) error {
	var teamID pgtype.UUID
	if participation.TeamID != nil {
		teamID = pgtype.UUID{
			Bytes: *participation.TeamID,
			Valid: true,
		}
	}

	err := r.Queries().UpsertParticipation(ctx, queries.UpsertParticipationParams{
		HackathonID:    participation.HackathonID,
		UserID:         participation.UserID,
		Status:         participation.Status,
		WishedRoleIds:  participation.WishedRoleIDs,
		MotivationText: pgtype.Text{String: participation.MotivationText, Valid: participation.MotivationText != ""},
		TeamID:         teamID,
		UpdatedAt:      participation.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to upsert participation: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *ParticipationRepository) Get(ctx context.Context, hackathonID, userID uuid.UUID) (*entity.Participation, error) {
	row, err := r.Queries().GetParticipation(ctx, queries.GetParticipationParams{
		HackathonID: hackathonID,
		UserID:      userID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get participation: %w", err)
	}

	var teamIDPtr *uuid.UUID
	if row.TeamID.Valid {
		teamID := uuid.UUID(row.TeamID.Bytes)
		teamIDPtr = &teamID
	}

	return &entity.Participation{
		HackathonID:    row.HackathonID,
		UserID:         row.UserID,
		Status:         row.Status,
		WishedRoleIDs:  row.WishedRoleIds,
		MotivationText: row.MotivationText.String,
		TeamID:         teamIDPtr,
		UpdatedAt:      row.UpdatedAt,
	}, nil
}

func (r *ParticipationRepository) ListByHackathon(ctx context.Context, hackathonID uuid.UUID) ([]*entity.Participation, error) {
	rows, err := r.Queries().ListParticipationsByHackathon(ctx, hackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to list participations: %w", pgxutil.MapDBError(err))
	}

	participations := make([]*entity.Participation, 0, len(rows))
	for _, row := range rows {
		var teamIDPtr *uuid.UUID
		if row.TeamID.Valid {
			teamID := uuid.UUID(row.TeamID.Bytes)
			teamIDPtr = &teamID
		}

		participations = append(participations, &entity.Participation{
			HackathonID:    row.HackathonID,
			UserID:         row.UserID,
			Status:         row.Status,
			WishedRoleIDs:  row.WishedRoleIds,
			MotivationText: row.MotivationText.String,
			TeamID:         teamIDPtr,
			UpdatedAt:      row.UpdatedAt,
		})
	}

	return participations, nil
}

func (r *ParticipationRepository) Delete(ctx context.Context, hackathonID, userID uuid.UUID) error {
	err := r.Queries().DeleteParticipation(ctx, queries.DeleteParticipationParams{
		HackathonID: hackathonID,
		UserID:      userID,
	})

	if err != nil {
		return fmt.Errorf("failed to delete participation: %w", pgxutil.MapDBError(err))
	}

	return nil
}
