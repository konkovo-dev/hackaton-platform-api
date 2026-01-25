package role

import (
	identityclient "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/client/identity"
)

type Service struct {
	uow            UnitOfWork
	roleRepo       StaffRoleRepository
	invitationRepo StaffInvitationRepository
	participRepo   ParticipationRepository
	identityClient *identityclient.Client
}

func NewService(
	uow UnitOfWork,
	roleRepo StaffRoleRepository,
	invitationRepo StaffInvitationRepository,
	participRepo ParticipationRepository,
	identityClient *identityclient.Client,
) *Service {
	return &Service{
		uow:            uow,
		roleRepo:       roleRepo,
		invitationRepo: invitationRepo,
		participRepo:   participRepo,
		identityClient: identityClient,
	}
}
