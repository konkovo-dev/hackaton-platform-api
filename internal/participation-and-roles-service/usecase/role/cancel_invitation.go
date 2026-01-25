package role

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	rolepolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/google/uuid"
)

type CancelInvitationIn struct {
	HackathonID  uuid.UUID
	InvitationID uuid.UUID
}

type CancelInvitationOut struct{}

type cancelPolicyRepositoryAdapter struct {
	roleRepo       StaffRoleRepository
	invitationRepo StaffInvitationRepository
}

func (a *cancelPolicyRepositoryAdapter) GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error) {
	return a.roleRepo.GetRoleStrings(ctx, hackathonID, userID)
}

func (a *cancelPolicyRepositoryAdapter) GetInvitationStatus(ctx context.Context, invitationID uuid.UUID) (string, uuid.UUID, error) {
	return a.invitationRepo.GetStatusAndHackathonID(ctx, invitationID)
}

func (s *Service) CancelInvitation(ctx context.Context, in CancelInvitationIn) (*CancelInvitationOut, error) {
	adapter := &cancelPolicyRepositoryAdapter{
		roleRepo:       s.roleRepo,
		invitationRepo: s.invitationRepo,
	}

	cancelPolicy := rolepolicy.NewCancelStaffInvitationPolicy(adapter)
	pctx, err := cancelPolicy.LoadContext(ctx, rolepolicy.CancelStaffInvitationParams{
		HackathonID:  in.HackathonID,
		InvitationID: in.InvitationID,
	})
	if err != nil {
		return nil, err
	}

	decision := cancelPolicy.Check(ctx, pctx)
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

	if invitation.HackathonID != in.HackathonID {
		return nil, ErrInvalidInput
	}

	// Check invitation status and provide specific error messages
	switch invitation.Status {
	case string(domain.InvitationStatusAccepted):
		return nil, fmt.Errorf("%w: cannot cancel accepted invitation", ErrConflict)
	case string(domain.InvitationStatusDeclined):
		return nil, fmt.Errorf("%w: cannot cancel rejected invitation", ErrConflict)
	case string(domain.InvitationStatusCanceled):
		return nil, fmt.Errorf("%w: invitation already canceled", ErrConflict)
	case string(domain.InvitationStatusExpired):
		return nil, fmt.Errorf("%w: cannot cancel expired invitation", ErrConflict)
	case string(domain.InvitationStatusPending):
		// OK to proceed
	default:
		return nil, fmt.Errorf("%w: invalid invitation status", ErrInvalidInput)
	}

	now := time.Now().UTC()
	err = s.invitationRepo.UpdateStatus(ctx, in.InvitationID, string(domain.InvitationStatusCanceled), now)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel invitation: %w", err)
	}

	return &CancelInvitationOut{}, nil
}
