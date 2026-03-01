package participation

type Service struct {
	participRepo ParticipationRepository
	roleRepo     StaffRoleRepository
	teamRoleRepo TeamRoleRepository
	outboxRepo   OutboxRepository
}

func NewService(
	participRepo ParticipationRepository,
	roleRepo StaffRoleRepository,
	teamRoleRepo TeamRoleRepository,
	outboxRepo OutboxRepository,
) *Service {
	return &Service{
		participRepo: participRepo,
		roleRepo:     roleRepo,
		teamRoleRepo: teamRoleRepo,
		outboxRepo:   outboxRepo,
	}
}
