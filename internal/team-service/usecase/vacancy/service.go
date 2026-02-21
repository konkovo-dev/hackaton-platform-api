package vacancy

import (
	"log/slog"
)

type Service struct {
	vacancyRepo    VacancyRepository
	teamRepo       TeamRepository
	membershipRepo MembershipRepository
	hackathonClient HackathonClient
	logger         *slog.Logger
}

func NewService(
	vacancyRepo VacancyRepository,
	teamRepo TeamRepository,
	membershipRepo MembershipRepository,
	hackathonClient HackathonClient,
	logger *slog.Logger,
) *Service {
	return &Service{
		vacancyRepo:     vacancyRepo,
		teamRepo:        teamRepo,
		membershipRepo:  membershipRepo,
		hackathonClient: hackathonClient,
		logger:          logger,
	}
}
