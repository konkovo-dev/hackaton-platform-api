package teammember

import (
	"log/slog"
)

type Service struct {
	membershipRepo MembershipRepository
	vacancyRepo    VacancyRepository
	teamRepo       TeamRepository
	txManager      TxManager
	hackathonClient HackathonClient
	parClient      ParticipationAndRolesClient
	logger         *slog.Logger
}

func NewService(
	membershipRepo MembershipRepository,
	vacancyRepo VacancyRepository,
	teamRepo TeamRepository,
	txManager TxManager,
	hackathonClient HackathonClient,
	parClient ParticipationAndRolesClient,
	logger *slog.Logger,
) *Service {
	return &Service{
		membershipRepo:  membershipRepo,
		vacancyRepo:     vacancyRepo,
		teamRepo:        teamRepo,
		txManager:       txManager,
		hackathonClient: hackathonClient,
		parClient:       parClient,
		logger:          logger,
	}
}
