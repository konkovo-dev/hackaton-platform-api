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
	db      queries.DBTX
}

func NewSkillRepository(db queries.DBTX) *SkillRepository {
	return &SkillRepository{
		queries: queries.New(db),
		db:      db,
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

type ListSkillCatalogParams struct {
	SearchQuery string
	Filters     []ListSkillCatalogFilter
	SortField   string
	SortDesc    bool
	Cursor      *ListSkillCatalogCursor
	Limit       int
}

type ListSkillCatalogFilter struct {
	Field     string
	Operation string
	Value     string
}

type ListSkillCatalogCursor struct {
	Name string
	ID   uuid.UUID
}

type ListSkillCatalogResult struct {
	ID   uuid.UUID
	Name string
}

func (r *SkillRepository) ListSkillCatalog(ctx context.Context, params ListSkillCatalogParams) ([]*ListSkillCatalogResult, bool, error) {
	query, args, err := r.buildListSkillCatalogQuery(params)
	if err != nil {
		return nil, false, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	results := make([]*ListSkillCatalogResult, 0)
	for rows.Next() {
		var id pgtype.UUID
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, false, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, &ListSkillCatalogResult{
			ID:   pgxutil.PgtypeToUUID(id),
			Name: name,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("rows error: %w", err)
	}

	hasMore := false
	if len(results) >= params.Limit {
		hasMore = true
	}

	return results, hasMore, nil
}

func (r *SkillRepository) buildListSkillCatalogQuery(params ListSkillCatalogParams) (string, []interface{}, error) {
	baseQuery := "SELECT sc.id, sc.name FROM identity.skill_catalog sc"

	whereClauses := []string{}
	args := []interface{}{}
	argCounter := 1

	if params.SearchQuery != "" {
		searchPattern := "%" + params.SearchQuery + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("sc.name ILIKE $%d", argCounter))
		args = append(args, searchPattern)
		argCounter++
	}

	for _, filter := range params.Filters {
		switch filter.Field {
		case "name":
			switch filter.Operation {
			case "CONTAINS":
				whereClauses = append(whereClauses, fmt.Sprintf("LOWER(sc.name) = LOWER($%d)", argCounter))
				args = append(args, filter.Value)
				argCounter++
			case "PREFIX":
				whereClauses = append(whereClauses, fmt.Sprintf("sc.name ILIKE $%d", argCounter))
				args = append(args, filter.Value+"%")
				argCounter++
			}
		}
	}

	if params.Cursor != nil {
		if params.SortDesc {
			whereClauses = append(whereClauses, fmt.Sprintf(
				"(sc.name, sc.id) < ($%d, $%d)",
				argCounter, argCounter+1,
			))
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf(
				"(sc.name, sc.id) > ($%d, $%d)",
				argCounter, argCounter+1,
			))
		}
		args = append(args, params.Cursor.Name, pgxutil.UUIDToPgtype(params.Cursor.ID))
		argCounter += 2
	}

	if len(whereClauses) > 0 {
		baseQuery += " WHERE " + whereClauses[0]
		for i := 1; i < len(whereClauses); i++ {
			baseQuery += " AND " + whereClauses[i]
		}
	}

	sortDirection := "ASC"
	if params.SortDesc {
		sortDirection = "DESC"
	}
	baseQuery += fmt.Sprintf(" ORDER BY sc.name %s, sc.id %s", sortDirection, sortDirection)

	baseQuery += fmt.Sprintf(" LIMIT $%d", argCounter)
	args = append(args, params.Limit)

	return baseQuery, args, nil
}
