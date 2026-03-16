package main

import (
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/repository/postgres"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/me"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/skills"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/users"
	outboxusecase "github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/outbox"
	authclient "github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/belikoooova/hackaton-platform-api/pkg/logger"
	natsclient "github.com/belikoooova/hackaton-platform-api/pkg/nats"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/belikoooova/hackaton-platform-api/pkg/s3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		logger.Module,
		authclient.Module,
		natsclient.Module,
		postgres.Module,
		s3.Module,
		idempotency.Module,
		me.Module,
		users.Module,
		skills.Module,
		outboxusecase.Module,
		outbox.Module,
		grpc.Module,
		fx.Provide(
			func(repo *postgres.UserRepository) me.UserRepository { return repo },
			func(repo *postgres.SkillRepository) me.SkillRepository { return repo },
			func(repo *postgres.ContactRepository) me.ContactRepository { return repo },
			func(repo *postgres.VisibilityRepository) me.VisibilityRepository { return repo },
			func(repo *postgres.AvatarUploadRepository) me.AvatarUploadRepository { return repo },
			func(pool *pgxpool.Pool) me.UnitOfWork {
				return pgxutil.NewUnitOfWork(pool, func(tx pgx.Tx) *me.TxRepositories {
					return &me.TxRepositories{
						Users:      postgres.NewUserRepository(tx),
						Skills:     postgres.NewSkillRepository(tx),
						Contacts:   postgres.NewContactRepository(tx),
						Visibility: postgres.NewVisibilityRepository(tx),
						Outbox:     postgres.NewOutboxRepository(tx),
					}
				})
			},
		),
		fx.Provide(
			func(repo *postgres.UserRepository) users.UserRepository { return repo },
			func(repo *postgres.SkillRepository) users.SkillRepository { return repo },
			func(repo *postgres.ContactRepository) users.ContactRepository { return repo },
			func(repo *postgres.VisibilityRepository) users.VisibilityRepository { return repo },
		),
		fx.Provide(
			func(repo *postgres.SkillRepository) skills.SkillRepository { return repo },
		),
	)

	app.Run()
}
