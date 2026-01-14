package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type SkillRepository struct {
	queries *queries.Queries
}

func NewSkillRepository(db queries.DBTX) *SkillRepository {
	return &SkillRepository{
		queries: queries.New(db),
	}
}

func (r *SkillRepository) ListCatalogSkillsByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.CatalogSkill, error) {
	pgIDs := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		pgIDs[i] = pgxutil.UUIDToPgtype(id)
	}

	rows, err := r.queries.SkillCatalogListByIDs(ctx, pgIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list catalog skills: %w", err)
	}

	skills := make([]*entity.CatalogSkill, 0, len(rows))
	for _, row := range rows {
		skills = append(skills, &entity.CatalogSkill{
			ID:   pgxutil.PgtypeToUUID(row.ID),
			Name: row.Name,
		})
	}

	return skills, nil
}

func (r *SkillRepository) GetUserCatalogSkills(ctx context.Context, userID uuid.UUID) ([]*entity.CatalogSkill, error) {
	rows, err := r.queries.SkillCatalogGetByUserID(ctx, pgxutil.UUIDToPgtype(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get user catalog skills: %w", err)
	}

	skills := make([]*entity.CatalogSkill, 0, len(rows))
	for _, row := range rows {
		skills = append(skills, &entity.CatalogSkill{
			ID:   pgxutil.PgtypeToUUID(row.ID),
			Name: row.Name,
		})
	}

	return skills, nil
}

func (r *SkillRepository) GetUserCustomSkills(ctx context.Context, userID uuid.UUID) ([]*entity.CustomSkill, error) {
	rows, err := r.queries.SkillCustomGetByUserID(ctx, pgxutil.UUIDToPgtype(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get user custom skills: %w", err)
	}

	skills := make([]*entity.CustomSkill, 0, len(rows))
	for _, row := range rows {
		skills = append(skills, &entity.CustomSkill{
			ID:     pgxutil.PgtypeToUUID(row.ID),
			UserID: pgxutil.PgtypeToUUID(row.UserID),
			Name:   row.Name,
		})
	}

	return skills, nil
}

func (r *SkillRepository) Update(ctx context.Context, userID uuid.UUID, catalogSkillIDs []uuid.UUID, customSkillNames []string) error {
	if err := r.queries.SkillCatalogDeleteByUserID(ctx, pgxutil.UUIDToPgtype(userID)); err != nil {
		return fmt.Errorf("failed to delete user catalog skills: %w", err)
	}

	if err := r.queries.SkillCustomDeleteByUserID(ctx, pgxutil.UUIDToPgtype(userID)); err != nil {
		return fmt.Errorf("failed to delete user custom skills: %w", err)
	}

	for _, catalogSkillID := range catalogSkillIDs {
		err := r.queries.SkillCatalogCreate(ctx, queries.SkillCatalogCreateParams{
			UserID:         pgxutil.UUIDToPgtype(userID),
			CatalogSkillID: pgxutil.UUIDToPgtype(catalogSkillID),
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("catalog skill not found: %s", catalogSkillID)
			}
			return fmt.Errorf("failed to create user catalog skill: %w", err)
		}
	}

	for _, customSkillName := range customSkillNames {
		err := r.queries.SkillCustomCreate(ctx, queries.SkillCustomCreateParams{
			ID:     pgxutil.UUIDToPgtype(uuid.New()),
			UserID: pgxutil.UUIDToPgtype(userID),
			Name:   customSkillName,
		})
		if err != nil {
			return fmt.Errorf("failed to create user custom skill: %w", err)
		}
	}

	return nil
}

func (r *SkillRepository) GetUsersCatalogSkills(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID][]*entity.CatalogSkill, error) {
	if len(userIDs) == 0 {
		return map[uuid.UUID][]*entity.CatalogSkill{}, nil
	}

	pgIDs := make([]pgtype.UUID, len(userIDs))
	for i, id := range userIDs {
		pgIDs[i] = pgxutil.UUIDToPgtype(id)
	}

	rows, err := r.queries.SkillCatalogGetByUserIDs(ctx, pgIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get users catalog skills: %w", err)
	}

	result := make(map[uuid.UUID][]*entity.CatalogSkill)
	for _, row := range rows {
		userID := pgxutil.PgtypeToUUID(row.UserID)
		skill := &entity.CatalogSkill{
			ID:   pgxutil.PgtypeToUUID(row.ID),
			Name: row.Name,
		}
		result[userID] = append(result[userID], skill)
	}

	return result, nil
}

func (r *SkillRepository) GetUsersCustomSkills(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID][]*entity.CustomSkill, error) {
	if len(userIDs) == 0 {
		return map[uuid.UUID][]*entity.CustomSkill{}, nil
	}

	pgIDs := make([]pgtype.UUID, len(userIDs))
	for i, id := range userIDs {
		pgIDs[i] = pgxutil.UUIDToPgtype(id)
	}

	rows, err := r.queries.SkillCustomGetByUserIDs(ctx, pgIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get users custom skills: %w", err)
	}

	result := make(map[uuid.UUID][]*entity.CustomSkill)
	for _, row := range rows {
		userID := pgxutil.PgtypeToUUID(row.UserID)
		skill := &entity.CustomSkill{
			ID:     pgxutil.PgtypeToUUID(row.ID),
			UserID: userID,
			Name:   row.Name,
		}
		result[userID] = append(result[userID], skill)
	}

	return result, nil
}
