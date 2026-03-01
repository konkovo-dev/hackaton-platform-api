package vacancy

import (
	"log/slog"
)

type Service struct {
	vacancyRepo     VacancyRepository
	teamRepo        TeamRepository
	membershipRepo  MembershipRepository
	outboxRepo      OutboxRepository
	hackathonClient HackathonClient
	logger          *slog.Logger
}

func NewService(
	vacancyRepo VacancyRepository,
	teamRepo TeamRepository,
	membershipRepo MembershipRepository,
	outboxRepo OutboxRepository,
	hackathonClient HackathonClient,
	logger *slog.Logger,
) *Service {
	return &Service{
		vacancyRepo:     vacancyRepo,
		teamRepo:        teamRepo,
		membershipRepo:  membershipRepo,
		outboxRepo:      outboxRepo,
		hackathonClient: hackathonClient,
		logger:          logger,
	}
}
