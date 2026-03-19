package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type VacancyRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewVacancyRepository(db queries.DBTX) *VacancyRepository {
	return &VacancyRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *VacancyRepository) GetByTeamIDs(ctx context.Context, teamIDs []uuid.UUID) ([]*entity.Vacancy, error) {
	if len(teamIDs) == 0 {
		return []*entity.Vacancy{}, nil
	}

	rows, err := r.Queries().GetVacanciesByTeamIDs(ctx, teamIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get vacancies by team ids: %w", pgxutil.MapDBError(err))
	}

	vacancies := make([]*entity.Vacancy, 0, len(rows))
	for _, row := range rows {
		vacancies = append(vacancies, &entity.Vacancy{
			ID:              row.ID,
			TeamID:          row.TeamID,
			Description:     row.Description,
			DesiredRoleIDs:  row.DesiredRoleIds,
			DesiredSkillIDs: row.DesiredSkillIds,
			SlotsTotal:      row.SlotsTotal,
			SlotsOpen:       row.SlotsOpen,
			IsSystem:        row.IsSystem,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}

	return vacancies, nil
}

func (r *VacancyRepository) GetByTeamID(ctx context.Context, teamID uuid.UUID) ([]*entity.Vacancy, error) {
	rows, err := r.Queries().GetVacanciesByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vacancies by team id: %w", pgxutil.MapDBError(err))
	}

	vacancies := make([]*entity.Vacancy, 0, len(rows))
	for _, row := range rows {
		vacancies = append(vacancies, &entity.Vacancy{
			ID:              row.ID,
			TeamID:          row.TeamID,
			Description:     row.Description,
			DesiredRoleIDs:  row.DesiredRoleIds,
			DesiredSkillIDs: row.DesiredSkillIds,
			SlotsTotal:      row.SlotsTotal,
			SlotsOpen:       row.SlotsOpen,
			IsSystem:        row.IsSystem,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}

	return vacancies, nil
}

func (r *VacancyRepository) GetByID(ctx context.Context, vacancyID uuid.UUID) (*entity.Vacancy, error) {
	row, err := r.Queries().GetVacancyByID(ctx, vacancyID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return nil, fmt.Errorf("failed to get vacancy by id: %w", err)
	}

	return &entity.Vacancy{
		ID:              row.ID,
		TeamID:          row.TeamID,
		Description:     row.Description,
		DesiredRoleIDs:  row.DesiredRoleIds,
		DesiredSkillIDs: row.DesiredSkillIds,
		SlotsTotal:      row.SlotsTotal,
		SlotsOpen:       row.SlotsOpen,
		IsSystem:        row.IsSystem,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}, nil
}

func (r *VacancyRepository) Create(ctx context.Context, vacancy *entity.Vacancy) error {
	row, err := r.Queries().CreateVacancy(ctx, queries.CreateVacancyParams{
		ID:              vacancy.ID,
		TeamID:          vacancy.TeamID,
		Description:     vacancy.Description,
		DesiredRoleIds:  vacancy.DesiredRoleIDs,
		DesiredSkillIds: vacancy.DesiredSkillIDs,
		SlotsTotal:      vacancy.SlotsTotal,
		SlotsOpen:       vacancy.SlotsOpen,
		IsSystem:        vacancy.IsSystem,
		CreatedAt:       vacancy.CreatedAt,
		UpdatedAt:       vacancy.UpdatedAt,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to create vacancy: %w", err)
	}

	vacancy.CreatedAt = row.CreatedAt
	vacancy.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *VacancyRepository) Update(ctx context.Context, vacancy *entity.Vacancy) error {
	row, err := r.Queries().UpdateVacancy(ctx, queries.UpdateVacancyParams{
		ID:              vacancy.ID,
		Description:     vacancy.Description,
		DesiredRoleIds:  vacancy.DesiredRoleIDs,
		DesiredSkillIds: vacancy.DesiredSkillIDs,
		SlotsTotal:      vacancy.SlotsTotal,
		SlotsOpen:       vacancy.SlotsOpen,
	})
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to update vacancy: %w", err)
	}

	vacancy.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *VacancyRepository) CountOccupiedSlots(ctx context.Context, vacancyID uuid.UUID) (int64, error) {
	count, err := r.Queries().CountOccupiedSlots(ctx, vacancyID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count occupied slots: %w", err)
	}

	return count, nil
}

func (r *VacancyRepository) CountTotalOpenSlots(ctx context.Context, teamID uuid.UUID) (int64, error) {
	result, err := r.Queries().CountTotalOpenSlots(ctx, teamID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return 0, fmt.Errorf("failed to count total open slots: %w", err)
	}

	return result, nil
}

func (r *VacancyRepository) DecrementSlotsOpen(ctx context.Context, vacancyID uuid.UUID) error {
	err := r.Queries().DecrementSlotsOpen(ctx, vacancyID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to decrement slots open: %w", err)
	}

	return nil
}

func (r *VacancyRepository) IncrementSlotsOpen(ctx context.Context, vacancyID uuid.UUID) error {
	err := r.Queries().IncrementSlotsOpen(ctx, vacancyID)
	if err != nil {
		err = pgxutil.MapDBError(err)
		return fmt.Errorf("failed to increment slots open: %w", err)
	}

	return nil
}
