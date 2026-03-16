package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/users"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/belikoooova/hackaton-platform-api/pkg/queryutil"
	"github.com/belikoooova/hackaton-platform-api/pkg/queryutil/sqlbuilder"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewUserRepository(db queries.DBTX) *UserRepository {
	return &UserRepository{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.Username = strings.ToLower(user.Username)

	err := r.Queries().UserCreate(ctx, queries.UserCreateParams{
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
		err = pgxutil.MapDBError(err)
		if pgxutil.IsConflict(err) {
			return fmt.Errorf("user already exists: %w", err)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	row, err := r.Queries().UserGetByID(ctx, pgxutil.UUIDToPgtype(id))
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, fmt.Errorf("user not found: %w", err)
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

	row, err := r.Queries().UserGetByUsername(ctx, username)
	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return nil, fmt.Errorf("user not found: %w", err)
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

	err := r.Queries().UserUpdate(ctx, queries.UserUpdateParams{
		ID:        pgxutil.UUIDToPgtype(user.ID),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarUrl: pgxutil.StringToPgtype(user.AvatarURL),
		Timezone:  user.Timezone,
		UpdatedAt: user.UpdatedAt,
	})

	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return fmt.Errorf("user not found: %w", err)
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

	rows, err := r.Queries().UserGetByIDs(ctx, pgIDs)
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

	rows, err := r.DB().Query(ctx, query, args...)
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
	qb := queryutil.NewQueryBuilder("SELECT DISTINCT u.id, u.username FROM identity.users u")

	qb.WithSearch([]string{"u.username", "u.first_name", "u.last_name"}, params.SearchQuery)

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
			qb.WithFilters(nonSkillGroups, params.FieldMapping)
		}
	}

	if params.Cursor != nil {
		qb.WithCursor([]queryutil.CursorField{
			{Column: "u.username", Value: params.Cursor.Username, Descending: false},
			{Column: "u.id", Value: pgxutil.UUIDToPgtype(params.Cursor.UserID), Descending: false},
		})
	}

	if len(params.Sort) > 0 {
		orderByFields := make([]queryutil.OrderByField, len(params.Sort))
		for i, sortField := range params.Sort {
			column := params.FieldMapping[sortField.Field]
			if column != "" {
				orderByFields[i] = queryutil.OrderByField{
					Column:     column,
					Descending: sortField.Direction == "DESC",
				}
			}
		}
		if len(orderByFields) > 0 {
			qb.WithOrderBy(orderByFields)
		}
	} else {
		qb.WithOrderBy([]queryutil.OrderByField{
			{Column: "u.username", Descending: false},
			{Column: "u.id", Descending: false},
		})
	}

	qb.WithLimit(params.Limit)

	query, args := qb.Build()
	return query, args, nil
}

func (r *UserRepository) UpdateAvatarURL(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	err := r.Queries().UserUpdateAvatarURL(ctx, queries.UserUpdateAvatarURLParams{
		ID:        pgxutil.UUIDToPgtype(userID),
		AvatarUrl: pgxutil.StringToPgtype(avatarURL),
	})

	if err != nil {
		err = pgxutil.MapDBError(err)
		if pgxutil.IsNotFound(err) {
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("failed to update avatar URL: %w", err)
	}

	return nil
}
