package participation

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	participationpolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type RegisterIn struct {
	HackathonID    uuid.UUID
	DesiredStatus  string
	WishedRoleIDs  []uuid.UUID
	MotivationText string
}

type RegisterOut struct {
	Participation *entity.Participation
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

func (s *Service) Register(ctx context.Context, in RegisterIn) (*RegisterOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if in.DesiredStatus != string(domain.ParticipationIndividual) &&
		in.DesiredStatus != string(domain.ParticipationLookingForTeam) {
		return nil, fmt.Errorf("%w: desired_status must be INDIVIDUAL_ACTIVE or LOOKING_FOR_TEAM", ErrInvalidInput)
	}

	adapter := &policyRepositoryAdapter{
		roleRepo:     s.roleRepo,
		participRepo: s.participRepo,
	}

	registerPolicy := participationpolicy.NewRegisterForHackathonPolicy(adapter)
	pctx, err := registerPolicy.LoadContext(ctx, participationpolicy.RegisterForHackathonParams{
		HackathonID:   in.HackathonID,
		DesiredStatus: in.DesiredStatus,
	})
	if err != nil {
		return nil, err
	}

	decision := registerPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	existing, err := s.participRepo.Get(ctx, in.HackathonID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing participation: %w", err)
	}
	if existing != nil && existing.Status != string(domain.ParticipationNone) {
		return nil, fmt.Errorf("%w: user is already registered for this hackathon", ErrConflict)
	}

	var wishedRoles []*entity.TeamRole
	if len(in.WishedRoleIDs) > 0 {
		wishedRoles, err = s.teamRoleRepo.GetByIDs(ctx, in.WishedRoleIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get team roles: %w", err)
		}
		if len(wishedRoles) != len(in.WishedRoleIDs) {
			return nil, fmt.Errorf("%w: some team role IDs are invalid", ErrInvalidInput)
		}
	}

	now := time.Now().UTC()

	participation := &entity.Participation{
		HackathonID:    in.HackathonID,
		UserID:         userUUID,
		Status:         in.DesiredStatus,
		TeamID:         nil,
		WishedRoles:    wishedRoles,
		MotivationText: in.MotivationText,
		RegisteredAt:   now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = s.participRepo.Create(ctx, participation)
	if err != nil {
		return nil, fmt.Errorf("failed to create participation: %w", err)
	}

	if len(in.WishedRoleIDs) > 0 {
		err = s.teamRoleRepo.SetForParticipation(ctx, in.HackathonID, userUUID, in.WishedRoleIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to set wished roles: %w", err)
		}
	}

	return &RegisterOut{
		Participation: participation,
	}, nil
}

func mapPolicyError(decision *policy.Decision) error {
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
