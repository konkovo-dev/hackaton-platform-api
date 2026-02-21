package main

import (
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/client/hackathon"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/client/participationroles"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/repository/postgres"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/transport/grpc"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/team"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teaminbox"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teammember"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/vacancy"
	authclient "github.com/belikoooova/hackaton-platform-api/pkg/auth/client"
	"github.com/belikoooova/hackaton-platform-api/pkg/logger"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		logger.Module,
		authclient.Module,
		postgres.Module,
		hackathon.Module,
		participationroles.Module,
		team.Module,
		vacancy.Module,
		teammember.Module,
		teaminbox.Module,
		grpc.Module,
		fx.Provide(
			func(repo *postgres.TeamRepository) team.TeamRepository { return repo },
			func(repo *postgres.VacancyRepository) team.VacancyRepository { return repo },
			func(repo *postgres.MembershipRepository) team.MembershipRepository { return repo },
			func(client *hackathon.Client) team.HackathonClient { return client },
			func(client *participationroles.Client) team.ParticipationAndRolesClient { return client },
			func(repo *postgres.VacancyRepository) vacancy.VacancyRepository { return repo },
			func(repo *postgres.TeamRepository) vacancy.TeamRepository { return repo },
			func(repo *postgres.MembershipRepository) vacancy.MembershipRepository { return repo },
			func(client *hackathon.Client) vacancy.HackathonClient { return client },
			func(repo *postgres.MembershipRepository) teammember.MembershipRepository { return repo },
			func(repo *postgres.VacancyRepository) teammember.VacancyRepository { return repo },
			func(repo *postgres.TeamRepository) teammember.TeamRepository { return repo },
			func(txMgr *pgxutil.TxManager) teammember.TxManager { return txMgr },
			func(client *hackathon.Client) teammember.HackathonClient { return client },
			func(client *participationroles.Client) teammember.ParticipationAndRolesClient { return client },
			func(repo *postgres.TeamInvitationRepository) teaminbox.TeamInvitationRepository { return repo },
			func(repo *postgres.JoinRequestRepository) teaminbox.JoinRequestRepository { return repo },
			func(repo *postgres.VacancyRepository) teaminbox.VacancyRepository { return repo },
			func(repo *postgres.TeamRepository) teaminbox.TeamRepository { return repo },
			func(repo *postgres.MembershipRepository) teaminbox.MembershipRepository { return repo },
			func(txMgr *pgxutil.TxManager) teaminbox.TxManager { return txMgr },
			func(client *hackathon.Client) teaminbox.HackathonClient { return client },
			func(client *participationroles.Client) teaminbox.ParticipationAndRolesClient { return client },
		),
	)

	app.Run()
}
