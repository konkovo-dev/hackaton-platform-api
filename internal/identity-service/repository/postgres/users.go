package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/repository/postgres/queries"
	pgxadapter "github.com/belikoooova/hackaton-platform-api/pkg/pgx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	queries *queries.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		queries: queries.New(pool),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.Username = strings.ToLower(user.Username)

	err := r.queries.CreateUser(ctx, queries.CreateUserParams{
		ID:        pgxadapter.UUIDToPgtype(user.ID),
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarUrl: pgxadapter.StringToPgtype(user.AvatarURL),
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
	row, err := r.queries.GetUserByID(ctx, pgxadapter.UUIDToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &entity.User{
		ID:        pgxadapter.PgtypeToUUID(row.ID),
		Username:  row.Username,
		FirstName: row.FirstName,
		LastName:  row.LastName,
		AvatarURL: pgxadapter.PgtypeToString(row.AvatarUrl),
		Timezone:  row.Timezone,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	username = strings.ToLower(username)

	row, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &entity.User{
		ID:        pgxadapter.PgtypeToUUID(row.ID),
		Username:  row.Username,
		FirstName: row.FirstName,
		LastName:  row.LastName,
		AvatarURL: pgxadapter.PgtypeToString(row.AvatarUrl),
		Timezone:  row.Timezone,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now().UTC()

	err := r.queries.UpdateUser(ctx, queries.UpdateUserParams{
		ID:        pgxadapter.UUIDToPgtype(user.ID),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarUrl: pgxadapter.StringToPgtype(user.AvatarURL),
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
