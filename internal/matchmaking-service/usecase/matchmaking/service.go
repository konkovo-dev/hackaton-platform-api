package matchmaking

import "log/slog"

type Service struct {
	userRepo         UserRepository
	participationRepo ParticipationRepository
	teamRepo         TeamRepository
	vacancyRepo      VacancyRepository
	hackathonClient  HackathonClient
	participationClient ParticipationRolesClient
	teamClient       TeamClient
	logger           *slog.Logger
}

func NewService(
	userRepo UserRepository,
	participationRepo ParticipationRepository,
	teamRepo TeamRepository,
	vacancyRepo VacancyRepository,
	hackathonClient HackathonClient,
	participationClient ParticipationRolesClient,
	teamClient TeamClient,
	logger *slog.Logger,
) *Service {
	return &Service{
		userRepo:         userRepo,
		participationRepo: participationRepo,
		teamRepo:         teamRepo,
		vacancyRepo:      vacancyRepo,
		hackathonClient:  hackathonClient,
		participationClient: participationClient,
		teamClient:       teamClient,
		logger:           logger,
	}
}
