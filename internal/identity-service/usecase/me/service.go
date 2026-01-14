package me

type Service struct {
	userRepo       UserRepository
	skillRepo      SkillRepository
	contactRepo    ContactRepository
	visibilityRepo VisibilityRepository
	uow            UnitOfWork
}

func NewService(
	userRepo UserRepository,
	skillRepo SkillRepository,
	contactRepo ContactRepository,
	visibilityRepo VisibilityRepository,
	uow UnitOfWork,
) *Service {
	return &Service{
		userRepo:       userRepo,
		skillRepo:      skillRepo,
		contactRepo:    contactRepo,
		visibilityRepo: visibilityRepo,
		uow:            uow,
	}
}
