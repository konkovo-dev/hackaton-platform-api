package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/users"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/belikoooova/hackaton-platform-api/pkg/queryutil/sqlbuilder"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	queries *queries.Queries
	db      queries.DBTX
}

func NewUserRepository(db queries.DBTX) *UserRepository {
	return &UserRepository{
		queries: queries.New(db),
		db:      db,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.Username = strings.ToLower(user.Username)

	err := r.queries.UserCreate(ctx, queries.UserCreateParams{
		ID:        pgxutil.UUIDToPgtype(user.ID),
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarUrl: pgxutil.StringToPgtype(user.AvatarURL),
		Timezone:  user.Timezone,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("user already exists: %w", err)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	row, err := r.queries.UserGetByID(ctx, pgxutil.UUIDToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &entity.User{
		ID:        pgxutil.PgtypeToUUID(row.ID),
		Username:  row.Username,
		FirstName: row.FirstName,
		LastName:  row.LastName,
		AvatarURL: pgxutil.PgtypeToString(row.AvatarUrl),
		Timezone:  row.Timezone,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	username = strings.ToLower(username)

	row, err := r.queries.UserGetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &entity.User{
		ID:        pgxutil.PgtypeToUUID(row.ID),
		Username:  row.Username,
		FirstName: row.FirstName,
		LastName:  row.LastName,
		AvatarURL: pgxutil.PgtypeToString(row.AvatarUrl),
		Timezone:  row.Timezone,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now().UTC()

	err := r.queries.UserUpdate(ctx, queries.UserUpdateParams{
		ID:        pgxutil.UUIDToPgtype(user.ID),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarUrl: pgxutil.StringToPgtype(user.AvatarURL),
		Timezone:  user.Timezone,
		UpdatedAt: user.UpdatedAt,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.User, error) {
	if len(ids) == 0 {
		return []*entity.User{}, nil
	}

	pgIDs := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		pgIDs[i] = pgxutil.UUIDToPgtype(id)
	}

	rows, err := r.queries.UserGetByIDs(ctx, pgIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by ids: %w", err)
	}

	users := make([]*entity.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, &entity.User{
			ID:        pgxutil.PgtypeToUUID(row.ID),
			Username:  row.Username,
			FirstName: row.FirstName,
			LastName:  row.LastName,
			AvatarURL: pgxutil.PgtypeToString(row.AvatarUrl),
			Timezone:  row.Timezone,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return users, nil
}

func (r *UserRepository) ListUsers(ctx context.Context, params users.ListUsersRepoParams) ([]*users.UserListResult, bool, error) {
	query, args, err := r.buildListUsersQuery(params)
	if err != nil {
		return nil, false, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	results := make([]*users.UserListResult, 0)
	for rows.Next() {
		var id pgtype.UUID
		var username string
		if err := rows.Scan(&id, &username); err != nil {
			return nil, false, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, &users.UserListResult{
			ID:       pgxutil.PgtypeToUUID(id),
			Username: username,
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

func (r *UserRepository) buildListUsersQuery(params users.ListUsersRepoParams) (string, []interface{}, error) {
	baseQuery := "SELECT DISTINCT u.id, u.username FROM identity.users u"

	whereClauses := []string{}
	args := []interface{}{}
	argCounter := 1

	if params.SearchQuery != "" {
		searchPattern := "%" + params.SearchQuery + "%"
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(u.username ILIKE $%d OR u.first_name ILIKE $%d OR u.last_name ILIKE $%d)",
			argCounter, argCounter+1, argCounter+2,
		))
		args = append(args, searchPattern, searchPattern, searchPattern)
		argCounter += 3
	}

	if params.Filters != nil && len(params.Filters.FilterGroups) > 0 {
		nonSkillGroups := []*sqlbuilder.FilterGroup{}
		for _, group := range params.Filters.FilterGroups {
			nonSkillFilters := []*sqlbuilder.Filter{}
			for _, filter := range group.Filters {
				if filter.Field != "skills" {
					nonSkillFilters = append(nonSkillFilters, filter)
				}
			}
			if len(nonSkillFilters) > 0 {
				nonSkillGroups = append(nonSkillGroups, &sqlbuilder.FilterGroup{Filters: nonSkillFilters})
			}
		}

		if len(nonSkillGroups) > 0 {
			wb := sqlbuilder.NewWhereBuilder(params.FieldMapping)
			wb.SetArgCounter(argCounter)

			whereClause := wb.Build(nonSkillGroups)
			if whereClause.SQL != "" {
				whereClauses = append(whereClauses, whereClause.SQL)
				args = append(args, whereClause.Args...)
				argCounter += len(whereClause.Args)
			}
		}
	}

	if params.Cursor != nil {
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(u.username, u.id) > ($%d, $%d)",
			argCounter, argCounter+1,
		))
		args = append(args, params.Cursor.Username, pgxutil.UUIDToPgtype(params.Cursor.UserID))
		argCounter += 2
	}

	if len(whereClauses) > 0 {
		baseQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	if len(params.Sort) > 0 {
		orderBy := sqlbuilder.BuildOrderBy(params.Sort, params.FieldMapping)
		if orderBy != "" {
			baseQuery += " ORDER BY " + orderBy
		}
	} else {
		baseQuery += " ORDER BY u.username ASC, u.id ASC"
	}

	baseQuery += fmt.Sprintf(" LIMIT $%d", argCounter)
	args = append(args, params.Limit)

	return baseQuery, args, nil
}
