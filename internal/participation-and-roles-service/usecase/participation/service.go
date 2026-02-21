package participation

type Service struct {
	participRepo  ParticipationRepository
	roleRepo      StaffRoleRepository
	teamRoleRepo  TeamRoleRepository
}

func NewService(
	participRepo ParticipationRepository,
	roleRepo StaffRoleRepository,
	teamRoleRepo TeamRoleRepository,
) *Service {
	return &Service{
		participRepo:  participRepo,
		roleRepo:      roleRepo,
		teamRoleRepo:  teamRoleRepo,
	}
}
