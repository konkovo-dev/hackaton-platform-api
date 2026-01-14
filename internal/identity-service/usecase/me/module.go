package me

import (
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/repository/postgres"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var Module = fx.Module("me-usecase",
	fx.Provide(
		NewService,
		NewUnitOfWork,
		NewUserRepository,
		NewSkillRepository,
		NewContactRepository,
		NewVisibilityRepository,
	),
)

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return postgres.NewUserRepository(pool)
}

func NewSkillRepository(pool *pgxpool.Pool) SkillRepository {
	return postgres.NewSkillRepository(pool)
}

func NewContactRepository(pool *pgxpool.Pool) ContactRepository {
	return postgres.NewContactRepository(pool)
}

func NewVisibilityRepository(pool *pgxpool.Pool) VisibilityRepository {
	return postgres.NewVisibilityRepository(pool)
}

func NewUnitOfWork(pool *pgxpool.Pool) UnitOfWork {
	factory := func(tx pgx.Tx) *TxRepositories {
		return &TxRepositories{
			Users:      postgres.NewUserRepository(tx),
			Skills:     postgres.NewSkillRepository(tx),
			Contacts:   postgres.NewContactRepository(tx),
			Visibility: postgres.NewVisibilityRepository(tx),
		}
	}

	return pgxutil.NewUnitOfWork(pool, factory)
}
