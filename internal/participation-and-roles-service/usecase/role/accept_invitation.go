package role

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	rolepolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/google/uuid"
)

type AcceptInvitationIn struct {
	InvitationID uuid.UUID
}

type AcceptInvitationOut struct{}

type acceptPolicyRepositoryAdapter struct {
	roleRepo       StaffRoleRepository
	invitationRepo StaffInvitationRepository
	participRepo   ParticipationRepository
}

func (a *acceptPolicyRepositoryAdapter) GetInvitationDetails(ctx context.Context, invitationID uuid.UUID) (bool, string, uuid.UUID, uuid.UUID, string, error) {
	return a.invitationRepo.GetDetails(ctx, invitationID)
}

func (a *acceptPolicyRepositoryAdapter) GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error) {
	return a.roleRepo.GetRoleStrings(ctx, hackathonID, userID)
}

func (a *acceptPolicyRepositoryAdapter) GetParticipationStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error) {
	return a.participRepo.GetStatus(ctx, hackathonID, userID)
}

func (s *Service) AcceptInvitation(ctx context.Context, in AcceptInvitationIn) (*AcceptInvitationOut, error) {
	adapter := &acceptPolicyRepositoryAdapter{
		roleRepo:       s.roleRepo,
		invitationRepo: s.invitationRepo,
		participRepo:   s.participRepo,
	}

	acceptPolicy := rolepolicy.NewAcceptStaffInvitationPolicy(adapter)
	pctx, err := acceptPolicy.LoadContext(ctx, rolepolicy.AcceptStaffInvitationParams{
		InvitationID: in.InvitationID,
	})
	if err != nil {
		return nil, err
	}

	decision := acceptPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, s.mapPolicyError(decision)
	}

	invitation, err := s.invitationRepo.GetByID(ctx, in.InvitationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}
	if invitation == nil {
		return nil, ErrInvalidInput
	}

	// Check invitation status and provide specific error messages
	switch invitation.Status {
	case string(domain.InvitationStatusAccepted):
		return nil, fmt.Errorf("%w: invitation already accepted", ErrConflict)
	case string(domain.InvitationStatusDeclined):
		return nil, fmt.Errorf("%w: invitation already rejected", ErrConflict)
	case string(domain.InvitationStatusCanceled):
		return nil, fmt.Errorf("%w: invitation has been canceled", ErrConflict)
	case string(domain.InvitationStatusExpired):
		return nil, fmt.Errorf("%w: invitation has expired", ErrConflict)
	case string(domain.InvitationStatusPending):
		// OK to proceed
	default:
		return nil, fmt.Errorf("%w: invalid invitation status", ErrInvalidInput)
	}

	now := time.Now().UTC()

	err = s.uow.Do(ctx, func(ctx context.Context, txRepos *TxRepositories) error {
		if err := txRepos.Invitations.UpdateStatus(ctx, in.InvitationID, string(domain.InvitationStatusAccepted), now); err != nil {
			return fmt.Errorf("failed to update invitation status: %w", err)
		}

		staffRole := &entity.StaffRole{
			HackathonID: invitation.HackathonID,
			UserID:      invitation.TargetUserID,
			Role:        invitation.RequestedRole,
			CreatedAt:   now,
		}

		if err := txRepos.Roles.Create(ctx, staffRole); err != nil {
			return fmt.Errorf("failed to assign role: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &AcceptInvitationOut{}, nil
}
