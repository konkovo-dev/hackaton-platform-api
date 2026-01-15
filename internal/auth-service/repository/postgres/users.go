package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/auth-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	queries *queries.Queries
}

func NewUserRepository(db queries.DBTX) *UserRepository {
	return &UserRepository{
		queries: queries.New(db),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.Username = strings.ToLower(user.Username)

	err := r.queries.CreateUser(ctx, queries.CreateUserParams{
		ID:        pgxutil.UUIDToPgtype(user.ID),
		Username:  user.Username,
		Email:     user.Email,
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
	row, err := r.queries.GetUserByID(ctx, pgxutil.UUIDToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &entity.User{
		ID:        pgxutil.PgtypeToUUID(row.ID),
		Username:  row.Username,
		Email:     row.Email,
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
		ID:        pgxutil.PgtypeToUUID(row.ID),
		Username:  row.Username,
		Email:     row.Email,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &entity.User{
		ID:        pgxutil.PgtypeToUUID(row.ID),
		Username:  row.Username,
		Email:     row.Email,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now().UTC()

	err := r.queries.UpdateUser(ctx, queries.UpdateUserParams{
		ID:        pgxutil.UUIDToPgtype(user.ID),
		Email:     user.Email,
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
