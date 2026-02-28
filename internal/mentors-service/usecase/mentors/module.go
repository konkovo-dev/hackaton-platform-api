package mentors

import (
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/client/hackathon"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/client/participationroles"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/client/team"
	"github.com/belikoooova/hackaton-platform-api/internal/mentors-service/repository/postgres"
	"go.uber.org/fx"
)

var Module = fx.Module("mentors-usecase",
	fx.Provide(
		func(tr *postgres.TicketRepository) TicketRepository {
			return tr
		},
		func(mr *postgres.MessageRepository) MessageRepository {
			return mr
		},
		func(hc *hackathon.Client) HackathonClient {
			return hc
		},
		func(prc *participationroles.Client) ParticipationRolesClient {
			return prc
		},
		func(tc *team.Client) TeamClient {
			return tc
		},
		NewService,
	),
)
