package postgres

import (
	"context"
	"fmt"
	"time"

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
		HackathonID:    row.HackathonID,
		UserID:         row.UserID,
		Status:         row.Status,
		TeamID:         teamID,
		WishedRoles:    nil,
		MotivationText: row.MotivationText,
		RegisteredAt:   row.RegisteredAt,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
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
		HackathonID:    participation.HackathonID,
		UserID:         participation.UserID,
		Status:         participation.Status,
		TeamID:         teamID,
		MotivationText: participation.MotivationText,
		RegisteredAt:   participation.RegisteredAt,
		CreatedAt:      participation.CreatedAt,
		UpdatedAt:      participation.UpdatedAt,
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

func (r *ParticipationRepository) UpdateProfile(ctx context.Context, hackathonID, userID uuid.UUID, motivationText string, updatedAt time.Time) error {
	err := r.Queries().UpdateParticipationProfile(ctx, queries.UpdateParticipationProfileParams{
		HackathonID:    hackathonID,
		UserID:         userID,
		MotivationText: motivationText,
		UpdatedAt:      updatedAt,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to update participation profile: %w", err)
	}
	return nil
}

func (r *ParticipationRepository) UpdateStatus(ctx context.Context, hackathonID, userID uuid.UUID, status string, updatedAt time.Time) error {
	err := r.Queries().UpdateParticipationStatus(ctx, queries.UpdateParticipationStatusParams{
		HackathonID: hackathonID,
		UserID:      userID,
		Status:      status,
		UpdatedAt:   updatedAt,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to update participation status: %w", err)
	}
	return nil
}

func (r *ParticipationRepository) Update(ctx context.Context, hackathonID, userID uuid.UUID, status string, teamID *uuid.UUID, updatedAt time.Time) error {
	var pgTeamID pgtype.UUID
	if teamID != nil {
		pgTeamID = pgtype.UUID{
			Bytes: *teamID,
			Valid: true,
		}
	}

	err := r.Queries().UpdateParticipation(ctx, queries.UpdateParticipationParams{
		HackathonID: hackathonID,
		Status:      status,
		TeamID:      pgTeamID,
		UpdatedAt:   updatedAt,
		UserID:      userID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to update participation: %w", err)
	}
	return nil
}

func (r *ParticipationRepository) Delete(ctx context.Context, hackathonID, userID uuid.UUID) error {
	err := r.Queries().DeleteParticipation(ctx, queries.DeleteParticipationParams{
		HackathonID: hackathonID,
		UserID:      userID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to delete participation: %w", err)
	}
	return nil
}

func (r *ParticipationRepository) GetHackathonIDsByUserParticipation(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	ids, err := r.Queries().GetHackathonIDsByUserParticipation(ctx, userID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get hackathon IDs by user participation: %w", err)
	}
	return ids, nil
}

func (r *ParticipationRepository) GetHackathonIDsByUserParticipationStatus(ctx context.Context, userID uuid.UUID, status string) ([]uuid.UUID, error) {
	ids, err := r.Queries().GetHackathonIDsByUserParticipationStatus(ctx, queries.GetHackathonIDsByUserParticipationStatusParams{
		UserID: userID,
		Status: status,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get hackathon IDs by user participation status: %w", err)
	}
	return ids, nil
}
