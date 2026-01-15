package users

type Service struct {
	userRepo       UserRepository
	skillRepo      SkillRepository
	contactRepo    ContactRepository
	visibilityRepo VisibilityRepository
}

func NewService(
	userRepo UserRepository,
	skillRepo SkillRepository,
	contactRepo ContactRepository,
	visibilityRepo VisibilityRepository,
) *Service {
	return &Service{
		userRepo:       userRepo,
		skillRepo:      skillRepo,
		contactRepo:    contactRepo,
		visibilityRepo: visibilityRepo,
	}
}
