package role

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	rolepolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/google/uuid"
)

type RejectInvitationIn struct {
	InvitationID uuid.UUID
}

type RejectInvitationOut struct{}

type rejectPolicyRepositoryAdapter struct {
	invitationRepo StaffInvitationRepository
}

func (a *rejectPolicyRepositoryAdapter) GetInvitationBasicInfo(ctx context.Context, invitationID uuid.UUID) (bool, string, uuid.UUID, error) {
	return a.invitationRepo.GetBasicInfo(ctx, invitationID)
}

func (s *Service) RejectInvitation(ctx context.Context, in RejectInvitationIn) (*RejectInvitationOut, error) {
	adapter := &rejectPolicyRepositoryAdapter{
		invitationRepo: s.invitationRepo,
	}

	rejectPolicy := rolepolicy.NewRejectStaffInvitationPolicy(adapter)
	pctx, err := rejectPolicy.LoadContext(ctx, rolepolicy.RejectStaffInvitationParams{
		InvitationID: in.InvitationID,
	})
	if err != nil {
		return nil, err
	}

	decision := rejectPolicy.Check(ctx, pctx)
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
	err = s.invitationRepo.UpdateStatus(ctx, in.InvitationID, string(domain.InvitationStatusDeclined), now)
	if err != nil {
		return nil, fmt.Errorf("failed to reject invitation: %w", err)
	}

	return &RejectInvitationOut{}, nil
}
