package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type TeamRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewTeamRepository(db queries.DBTX) *TeamRepository {
	return &TeamRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *TeamRepository) List(ctx context.Context, hackathonID uuid.UUID, limit, offset int32) ([]*entity.Team, error) {
	rows, err := r.Queries().ListTeams(ctx, queries.ListTeamsParams{
		HackathonID: hackathonID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", pgxutil.MapDBError(err))
	}

	teams := make([]*entity.Team, 0, len(rows))
	for _, row := range rows {
		teams = append(teams, &entity.Team{
			ID:          row.ID,
			HackathonID: row.HackathonID,
			Name:        row.Name,
			Description: row.Description,
			IsJoinable:  row.IsJoinable,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}

	return teams, nil
}

func (r *TeamRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Team, error) {
	row, err := r.Queries().GetTeamByID(ctx, id)
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, fmt.Errorf("team not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return &entity.Team{
		ID:          row.ID,
		HackathonID: row.HackathonID,
		Name:        row.Name,
		Description: row.Description,
		IsJoinable:  row.IsJoinable,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func (r *TeamRepository) GetByIDAndHackathonID(ctx context.Context, id, hackathonID uuid.UUID) (*entity.Team, error) {
	row, err := r.Queries().GetTeamByIDAndHackathonID(ctx, queries.GetTeamByIDAndHackathonIDParams{
		ID:          id,
		HackathonID: hackathonID,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, fmt.Errorf("team not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return &entity.Team{
		ID:          row.ID,
		HackathonID: row.HackathonID,
		Name:        row.Name,
		Description: row.Description,
		IsJoinable:  row.IsJoinable,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func (r *TeamRepository) Create(ctx context.Context, team *entity.Team) error {
	row, err := r.Queries().CreateTeam(ctx, queries.CreateTeamParams{
		ID:          team.ID,
		HackathonID: team.HackathonID,
		Name:        team.Name,
		Description: team.Description,
		IsJoinable:  team.IsJoinable,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to create team: %w", err)
	}

	team.CreatedAt = row.CreatedAt
	team.UpdatedAt = row.UpdatedAt

	return nil
}

func (r *TeamRepository) CheckNameExists(ctx context.Context, hackathonID uuid.UUID, name string) (bool, error) {
	exists, err := r.Queries().CheckTeamNameExists(ctx, queries.CheckTeamNameExistsParams{
		HackathonID: hackathonID,
		Name:        name,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return false, fmt.Errorf("failed to check team name: %w", err)
	}

	return exists, nil
}

func (r *TeamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.Queries().DeleteTeam(ctx, id)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to delete team: %w", err)
	}

	return nil
}

func (r *TeamRepository) Update(ctx context.Context, team *entity.Team) error {
	row, err := r.Queries().UpdateTeam(ctx, queries.UpdateTeamParams{
		ID:          team.ID,
		Name:        team.Name,
		Description: team.Description,
		IsJoinable:  team.IsJoinable,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to update team: %w", err)
	}

	team.UpdatedAt = row.UpdatedAt

	return nil
}
