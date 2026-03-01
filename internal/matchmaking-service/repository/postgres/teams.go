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

type TeamRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewTeamRepository(db queries.DBTX) *TeamRepository {
	return &TeamRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *TeamRepository) Upsert(ctx context.Context, team *entity.Team) error {
	err := r.Queries().UpsertTeam(ctx, queries.UpsertTeamParams{
		TeamID:      team.TeamID,
		HackathonID: team.HackathonID,
		Name:        team.Name,
		Description: pgtype.Text{String: team.Description, Valid: team.Description != ""},
		IsJoinable:  team.IsJoinable,
		UpdatedAt:   team.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to upsert team: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *TeamRepository) GetByID(ctx context.Context, teamID uuid.UUID) (*entity.Team, error) {
	row, err := r.Queries().GetTeamByID(ctx, teamID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return &entity.Team{
		TeamID:      row.TeamID,
		HackathonID: row.HackathonID,
		Name:        row.Name,
		Description: row.Description.String,
		IsJoinable:  row.IsJoinable,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func (r *TeamRepository) ListByHackathon(ctx context.Context, hackathonID uuid.UUID) ([]*entity.Team, error) {
	rows, err := r.Queries().ListTeamsByHackathon(ctx, hackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", pgxutil.MapDBError(err))
	}

	teams := make([]*entity.Team, 0, len(rows))
	for _, row := range rows {
		teams = append(teams, &entity.Team{
			TeamID:      row.TeamID,
			HackathonID: row.HackathonID,
			Name:        row.Name,
			Description: row.Description.String,
			IsJoinable:  row.IsJoinable,
			UpdatedAt:   row.UpdatedAt,
		})
	}

	return teams, nil
}

func (r *TeamRepository) Delete(ctx context.Context, teamID uuid.UUID) error {
	err := r.Queries().DeleteTeam(ctx, teamID)

	if err != nil {
		return fmt.Errorf("failed to delete team: %w", pgxutil.MapDBError(err))
	}

	return nil
}
