package main

import (
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/client/hackathon"
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/client/participationroles"
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/client/team"
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/repository/postgres"
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/transport/grpc"
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/usecase/matchmaking"
	syncusecase "github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/usecase/sync"
	authclient "github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/belikoooova/hackaton-platform-api/pkg/logger"
	natsclient "github.com/belikoooova/hackaton-platform-api/pkg/nats"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		logger.Module,
		authclient.Module,
		idempotency.Module,
		natsclient.Module,
		postgres.Module,
		hackathon.Module,
		participationroles.Module,
		team.Module,
		matchmaking.Module,
		syncusecase.Module,
		grpc.Module,
		fx.Provide(
			func(r *postgres.UserRepository) matchmaking.UserRepository {
				return r
			},
			func(r *postgres.ParticipationRepository) matchmaking.ParticipationRepository {
				return r
			},
			func(r *postgres.TeamRepository) matchmaking.TeamRepository {
				return r
			},
			func(r *postgres.VacancyRepository) matchmaking.VacancyRepository {
				return r
			},
			func(c *hackathon.Client) matchmaking.HackathonClient {
				return c
			},
			func(c *participationroles.Client) matchmaking.ParticipationRolesClient {
				return c
			},
			func(c *team.Client) matchmaking.TeamClient {
				return c
			},
			func(r *postgres.UserRepository) syncusecase.UserRepository {
				return r
			},
			func(r *postgres.ParticipationRepository) syncusecase.ParticipationRepository {
				return r
			},
			func(r *postgres.TeamRepository) syncusecase.TeamRepository {
				return r
			},
			func(r *postgres.VacancyRepository) syncusecase.VacancyRepository {
				return r
			},
		),
	)

	app.Run()
}
