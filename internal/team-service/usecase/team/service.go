package team

import (
	"log/slog"

	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
)

type Service struct {
	teamRepo        TeamRepository
	vacancyRepo     VacancyRepository
	membershipRepo  MembershipRepository
	outboxRepo      OutboxRepository
	txManager       *pgxutil.TxManager
	hackathonClient HackathonClient
	parClient       ParticipationAndRolesClient
	logger          *slog.Logger
}

func NewService(
	teamRepo TeamRepository,
	vacancyRepo VacancyRepository,
	membershipRepo MembershipRepository,
	outboxRepo OutboxRepository,
	txManager *pgxutil.TxManager,
	hackathonClient HackathonClient,
	parClient ParticipationAndRolesClient,
	logger *slog.Logger,
) *Service {
	return &Service{
		teamRepo:        teamRepo,
		vacancyRepo:     vacancyRepo,
		membershipRepo:  membershipRepo,
		outboxRepo:      outboxRepo,
		txManager:       txManager,
		hackathonClient: hackathonClient,
		parClient:       parClient,
		logger:          logger,
	}
}
