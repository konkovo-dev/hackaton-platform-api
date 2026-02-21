package teaminbox

import (
	"log/slog"
)

type Service struct {
	invitationRepo    TeamInvitationRepository
	joinRequestRepo   JoinRequestRepository
	vacancyRepo       VacancyRepository
	teamRepo          TeamRepository
	membershipRepo    MembershipRepository
	txManager         TxManager
	hackathonClient   HackathonClient
	parClient         ParticipationAndRolesClient
	logger            *slog.Logger
}

func NewService(
	invitationRepo TeamInvitationRepository,
	joinRequestRepo JoinRequestRepository,
	vacancyRepo VacancyRepository,
	teamRepo TeamRepository,
	membershipRepo MembershipRepository,
	txManager TxManager,
	hackathonClient HackathonClient,
	parClient ParticipationAndRolesClient,
	logger *slog.Logger,
) *Service {
	return &Service{
		invitationRepo:  invitationRepo,
		joinRequestRepo: joinRequestRepo,
		vacancyRepo:     vacancyRepo,
		teamRepo:        teamRepo,
		membershipRepo:  membershipRepo,
		txManager:       txManager,
		hackathonClient: hackathonClient,
		parClient:       parClient,
		logger:          logger,
	}
}
