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
	),
)

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
