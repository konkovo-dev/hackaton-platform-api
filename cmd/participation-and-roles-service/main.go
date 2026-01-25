package main

import (
	identityclient "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/client/identity"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/repository/postgres"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/transport/grpc"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/role"
	authclient "github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/belikoooova/hackaton-platform-api/pkg/logger"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		logger.Module,
		authclient.Module,
		identityclient.Module,
		postgres.Module,
		idempotency.Module,
		role.Module,
		grpc.Module,
		fx.Provide(
			func(repo *postgres.StaffRoleRepository) role.StaffRoleRepository { return repo },
			func(repo *postgres.StaffInvitationRepository) role.StaffInvitationRepository { return repo },
			func(repo *postgres.ParticipationRepository) role.ParticipationRepository { return repo },
			func(pool *pgxpool.Pool) role.UnitOfWork {
				return pgxutil.NewUnitOfWork(pool, func(tx pgx.Tx) *role.TxRepositories {
					return &role.TxRepositories{
						Roles:       postgres.NewStaffRoleRepository(tx),
						Invitations: postgres.NewStaffInvitationRepository(tx),
					}
				})
			},
		),
	)

	app.Run()
}
