package role

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	rolepolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

const InvitationExpirationDays = 7

type CreateInvitationIn struct {
	HackathonID   uuid.UUID
	TargetUserID  uuid.UUID
	RequestedRole string
	Message       string
}

type CreateInvitationOut struct {
	InvitationID uuid.UUID
}

type policyRepositoryAdapter struct {
	roleRepo     StaffRoleRepository
	participRepo ParticipationRepository
}

func (a *policyRepositoryAdapter) GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error) {
	return a.roleRepo.GetRoleStrings(ctx, hackathonID, userID)
}

func (a *policyRepositoryAdapter) GetParticipationStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error) {
	return a.participRepo.GetStatus(ctx, hackathonID, userID)
}

func (s *Service) CreateInvitation(ctx context.Context, in CreateInvitationIn) (*CreateInvitationOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if in.TargetUserID == userUUID {
		return nil, fmt.Errorf("%w: cannot invite yourself", ErrInvalidInput)
	}

	if in.RequestedRole != string(domain.RoleOrganizer) &&
		in.RequestedRole != string(domain.RoleMentor) &&
		in.RequestedRole != string(domain.RoleJudge) {
		return nil, fmt.Errorf("%w: invalid role", ErrInvalidInput)
	}

	// Check if target user exists
	exists, err := s.identityClient.UserExists(ctx, in.TargetUserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("%w: user not found", ErrNotFound)
	}

	// Check if target user is already a staff member
	existingRoles, err := s.roleRepo.GetRoleStrings(ctx, in.HackathonID, in.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing roles: %w", err)
	}
	if len(existingRoles) > 0 {
		return nil, fmt.Errorf("%w: user is already a staff member", ErrConflict)
	}

	adapter := &policyRepositoryAdapter{
		roleRepo:     s.roleRepo,
		participRepo: s.participRepo,
	}

	createPolicy := rolepolicy.NewCreateStaffInvitationPolicy(adapter)
	pctx, err := createPolicy.LoadContext(ctx, rolepolicy.CreateStaffInvitationParams{
		HackathonID:   in.HackathonID,
		TargetUserID:  in.TargetUserID,
		RequestedRole: in.RequestedRole,
	})
	if err != nil {
		return nil, err
	}

	decision := createPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, s.mapPolicyError(decision)
	}

	existingInvitation, err := s.invitationRepo.GetPendingInvitationForUser(ctx, in.HackathonID, in.TargetUserID, in.RequestedRole)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing invitation: %w", err)
	}
	if existingInvitation != nil {
		return nil, fmt.Errorf("%w: pending invitation already exists", ErrConflict)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(InvitationExpirationDays * 24 * time.Hour)

	invitation := &entity.StaffInvitation{
		ID:              uuid.New(),
		HackathonID:     in.HackathonID,
		TargetUserID:    in.TargetUserID,
		RequestedRole:   in.RequestedRole,
		CreatedByUserID: userUUID,
		Message:         in.Message,
		Status:          string(domain.InvitationStatusPending),
		CreatedAt:       now,
		UpdatedAt:       now,
		ExpiresAt:       &expiresAt,
	}

	err = s.invitationRepo.Create(ctx, invitation)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	return &CreateInvitationOut{
		InvitationID: invitation.ID,
	}, nil
}

func (s *Service) mapPolicyError(decision *policy.Decision) error {
	if len(decision.Violations) == 0 {
		return ErrUnauthorized
	}

	v := decision.Violations[0]
	switch v.Code {
	case policy.ViolationCodeForbidden:
		return fmt.Errorf("%w: %s", ErrForbidden, v.Message)
	case policy.ViolationCodeNotFound:
		return fmt.Errorf("%w: %s", ErrNotFound, v.Message)
	case policy.ViolationCodePolicyRule:
		return fmt.Errorf("%w: %s", ErrInvalidInput, v.Message)
	case policy.ViolationCodeConflict:
		return fmt.Errorf("%w: %s", ErrConflict, v.Message)
	default:
		return ErrUnauthorized
	}
}
