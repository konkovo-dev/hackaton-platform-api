package main

import (
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/client/participationroles"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/repository/postgres"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/transport/grpc"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/announcement"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/hackathon"
	outboxHandlers "github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/outbox"
	authclient "github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/belikoooova/hackaton-platform-api/pkg/logger"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		logger.Module,
		authclient.Module,
		postgres.Module,
		participationroles.Module,
		idempotency.Module,
		hackathon.Module,
		announcement.Module,
		outboxHandlers.Module,
		outbox.Module,
		grpc.Module,
		fx.Provide(
			func(repo *postgres.HackathonRepository) hackathon.HackathonRepository { return repo },
			func(repo *postgres.HackathonLinkRepository) hackathon.HackathonLinkRepository { return repo },
			func(repo *postgres.HackathonRepository) announcement.HackathonRepository { return repo },
			func(repo *postgres.AnnouncementRepository) announcement.AnnouncementRepository { return repo },
			func(client *participationroles.Client) hackathon.ParticipationAndRolesClient { return client },
			func(client *participationroles.Client) announcement.ParticipationAndRolesClient { return client },
			func(pool *pgxpool.Pool) hackathon.UnitOfWork {
				return pgxutil.NewUnitOfWork(pool, func(tx pgx.Tx) *hackathon.TxRepositories {
					return &hackathon.TxRepositories{
						Hackathons: postgres.NewHackathonRepository(tx),
						Links:      postgres.NewHackathonLinkRepository(tx),
						Outbox:     postgres.NewOutboxRepository(tx),
					}
				})
			},
		),
	)

	app.Run()
}
