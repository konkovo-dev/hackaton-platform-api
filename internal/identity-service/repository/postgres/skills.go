package postgres

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/belikoooova/hackaton-platform-api/pkg/queryutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type SkillRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewSkillRepository(db queries.DBTX) *SkillRepository {
	return &SkillRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *SkillRepository) ListCatalogSkillsByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.CatalogSkill, error) {
	pgIDs := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		pgIDs[i] = pgxutil.UUIDToPgtype(id)
	}

	rows, err := r.Queries().SkillCatalogListByIDs(ctx, pgIDs)
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
	rows, err := r.Queries().SkillCatalogGetByUserID(ctx, pgxutil.UUIDToPgtype(userID))
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
	rows, err := r.Queries().SkillCustomGetByUserID(ctx, pgxutil.UUIDToPgtype(userID))
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
	if err := r.Queries().SkillCatalogDeleteByUserID(ctx, pgxutil.UUIDToPgtype(userID)); err != nil {
		return fmt.Errorf("failed to delete user catalog skills: %w", err)
	}

	if err := r.Queries().SkillCustomDeleteByUserID(ctx, pgxutil.UUIDToPgtype(userID)); err != nil {
		return fmt.Errorf("failed to delete user custom skills: %w", err)
	}

	for _, catalogSkillID := range catalogSkillIDs {
		err := r.Queries().SkillCatalogCreate(ctx, queries.SkillCatalogCreateParams{
			UserID:         pgxutil.UUIDToPgtype(userID),
			CatalogSkillID: pgxutil.UUIDToPgtype(catalogSkillID),
		})
		if err != nil {
			err = pgxutil.MapDBError(err)
			if pgxutil.IsNotFound(err) {
				return fmt.Errorf("catalog skill not found: %s: %w", catalogSkillID, err)
			}
			return fmt.Errorf("failed to create user catalog skill: %w", err)
		}
	}

	for _, customSkillName := range customSkillNames {
		err := r.Queries().SkillCustomCreate(ctx, queries.SkillCustomCreateParams{
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

	rows, err := r.Queries().SkillCatalogGetByUserIDs(ctx, pgIDs)
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

	rows, err := r.Queries().SkillCustomGetByUserIDs(ctx, pgIDs)
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

	rows, err := r.DB().Query(ctx, query, args...)
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
	qb := queryutil.NewQueryBuilder("SELECT sc.id, sc.name FROM identity.skill_catalog sc")

	qb.WithSearch([]string{"sc.name"}, params.SearchQuery)

	for _, filter := range params.Filters {
		switch filter.Field {
		case "name":
			switch filter.Operation {
			case "CONTAINS":
				qb.WithCustomWhere("LOWER(sc.name) = LOWER(?)", filter.Value)
			case "PREFIX":
				qb.WithCustomWhere("sc.name ILIKE ?", filter.Value+"%")
			}
		}
	}

	if params.Cursor != nil {
		qb.WithCursor([]queryutil.CursorField{
			{Column: "sc.name", Value: params.Cursor.Name, Descending: params.SortDesc},
			{Column: "sc.id", Value: pgxutil.UUIDToPgtype(params.Cursor.ID), Descending: params.SortDesc},
		})
	}

	qb.WithOrderBy([]queryutil.OrderByField{
		{Column: "sc.name", Descending: params.SortDesc},
		{Column: "sc.id", Descending: params.SortDesc},
	})

	qb.WithLimit(params.Limit)

	query, args := qb.Build()
	return query, args, nil
}
