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

type VacancyRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewVacancyRepository(db queries.DBTX) *VacancyRepository {
	return &VacancyRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *VacancyRepository) Upsert(ctx context.Context, vacancy *entity.Vacancy) error {
	err := r.Queries().UpsertVacancy(ctx, queries.UpsertVacancyParams{
		VacancyID:       vacancy.VacancyID,
		TeamID:          vacancy.TeamID,
		HackathonID:     vacancy.HackathonID,
		Description:     pgtype.Text{String: vacancy.Description, Valid: vacancy.Description != ""},
		DesiredRoleIds:  vacancy.DesiredRoleIDs,
		DesiredSkillIds: vacancy.DesiredSkillIDs,
		SlotsOpen:       vacancy.SlotsOpen,
		UpdatedAt:       vacancy.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to upsert vacancy: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *VacancyRepository) GetByID(ctx context.Context, vacancyID uuid.UUID) (*entity.Vacancy, error) {
	row, err := r.Queries().GetVacancyByID(ctx, vacancyID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get vacancy: %w", err)
	}

	return &entity.Vacancy{
		VacancyID:       row.VacancyID,
		TeamID:          row.TeamID,
		HackathonID:     row.HackathonID,
		Description:     row.Description.String,
		DesiredRoleIDs:  row.DesiredRoleIds,
		DesiredSkillIDs: row.DesiredSkillIds,
		SlotsOpen:       row.SlotsOpen,
		UpdatedAt:       row.UpdatedAt,
	}, nil
}

func (r *VacancyRepository) GetByTeamID(ctx context.Context, teamID uuid.UUID) ([]*entity.Vacancy, error) {
	rows, err := r.Queries().GetVacanciesByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vacancies by team: %w", pgxutil.MapDBError(err))
	}

	vacancies := make([]*entity.Vacancy, 0, len(rows))
	for _, row := range rows {
		vacancies = append(vacancies, &entity.Vacancy{
			VacancyID:       row.VacancyID,
			TeamID:          row.TeamID,
			HackathonID:     row.HackathonID,
			Description:     row.Description.String,
			DesiredRoleIDs:  row.DesiredRoleIds,
			DesiredSkillIDs: row.DesiredSkillIds,
			SlotsOpen:       row.SlotsOpen,
			UpdatedAt:       row.UpdatedAt,
		})
	}

	return vacancies, nil
}

func (r *VacancyRepository) ListByHackathon(ctx context.Context, hackathonID uuid.UUID) ([]*entity.Vacancy, error) {
	rows, err := r.Queries().ListVacanciesByHackathon(ctx, hackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to list vacancies: %w", pgxutil.MapDBError(err))
	}

	vacancies := make([]*entity.Vacancy, 0, len(rows))
	for _, row := range rows {
		vacancies = append(vacancies, &entity.Vacancy{
			VacancyID:       row.VacancyID,
			TeamID:          row.TeamID,
			HackathonID:     row.HackathonID,
			Description:     row.Description.String,
			DesiredRoleIDs:  row.DesiredRoleIds,
			DesiredSkillIDs: row.DesiredSkillIds,
			SlotsOpen:       row.SlotsOpen,
			UpdatedAt:       row.UpdatedAt,
		})
	}

	return vacancies, nil
}

func (r *VacancyRepository) Delete(ctx context.Context, vacancyID uuid.UUID) error {
	err := r.Queries().DeleteVacancy(ctx, vacancyID)

	if err != nil {
		return fmt.Errorf("failed to delete vacancy: %w", pgxutil.MapDBError(err))
	}

	return nil
}

func (r *VacancyRepository) DeleteByTeamID(ctx context.Context, teamID uuid.UUID) error {
	err := r.Queries().DeleteVacanciesByTeamID(ctx, teamID)

	if err != nil {
		return fmt.Errorf("failed to delete vacancies by team: %w", pgxutil.MapDBError(err))
	}

	return nil
}
