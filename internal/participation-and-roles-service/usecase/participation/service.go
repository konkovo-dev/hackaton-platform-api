package participation

type Service struct {
	participRepo    ParticipationRepository
	roleRepo        StaffRoleRepository
	teamRoleRepo    TeamRoleRepository
	outboxRepo      OutboxRepository
	hackathonClient HackathonClient
}

func NewService(
	participRepo ParticipationRepository,
	roleRepo StaffRoleRepository,
	teamRoleRepo TeamRoleRepository,
	outboxRepo OutboxRepository,
	hackathonClient HackathonClient,
) *Service {
	return &Service{
		participRepo:    participRepo,
		roleRepo:        roleRepo,
		teamRoleRepo:    teamRoleRepo,
		outboxRepo:      outboxRepo,
		hackathonClient: hackathonClient,
	}
}
