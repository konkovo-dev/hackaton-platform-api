package hackathon

type Service struct {
	uow           UnitOfWork
	hackathonRepo HackathonRepository
	linkRepo      HackathonLinkRepository
	parClient     ParticipationAndRolesClient
}

func NewService(
	uow UnitOfWork,
	hackathonRepo HackathonRepository,
	linkRepo HackathonLinkRepository,
	parClient ParticipationAndRolesClient,
) *Service {
	return &Service{
		uow:           uow,
		hackathonRepo: hackathonRepo,
		linkRepo:      linkRepo,
		parClient:     parClient,
	}
}
