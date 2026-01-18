package hackathon

type Service struct {
	uow           UnitOfWork
	hackathonRepo HackathonRepository
	linkRepo      HackathonLinkRepository
}

func NewService(
	uow UnitOfWork,
	hackathonRepo HackathonRepository,
	linkRepo HackathonLinkRepository,
) *Service {
	return &Service{
		uow:           uow,
		hackathonRepo: hackathonRepo,
		linkRepo:      linkRepo,
	}
}
