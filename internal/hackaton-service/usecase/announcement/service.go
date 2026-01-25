package announcement

type Service struct {
	hackathonRepo    HackathonRepository
	announcementRepo AnnouncementRepository
	parClient        ParticipationAndRolesClient
}

func NewService(
	hackathonRepo HackathonRepository,
	announcementRepo AnnouncementRepository,
	parClient ParticipationAndRolesClient,
) *Service {
	return &Service{
		hackathonRepo:    hackathonRepo,
		announcementRepo: announcementRepo,
		parClient:        parClient,
	}
}
