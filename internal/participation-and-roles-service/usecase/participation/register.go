package participation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain/entity"
	participationpolicy "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/policy"
	outboxusecase "github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
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
	roleRepo        StaffRoleRepository
	participRepo    ParticipationRepository
	hackathonClient HackathonClient
}

func (a *policyRepositoryAdapter) GetRoleStrings(ctx context.Context, hackathonID, userID uuid.UUID) ([]string, error) {
	return a.roleRepo.GetRoleStrings(ctx, hackathonID, userID)
}

func (a *policyRepositoryAdapter) GetParticipationStatus(ctx context.Context, hackathonID, userID uuid.UUID) (string, error) {
	return a.participRepo.GetStatus(ctx, hackathonID, userID)
}

func (a *policyRepositoryAdapter) GetHackathonInfo(ctx context.Context, hackathonID uuid.UUID) (*participationpolicy.HackathonInfo, error) {
	info, err := a.hackathonClient.GetHackathonInfo(ctx, hackathonID)
	if err != nil {
		return nil, err
	}
	return &participationpolicy.HackathonInfo{
		Stage:           info.Stage,
		AllowIndividual: info.AllowIndividual,
		AllowTeam:       info.AllowTeam,
	}, nil
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
		roleRepo:        s.roleRepo,
		participRepo:    s.participRepo,
		hackathonClient: s.hackathonClient,
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

	wishedRoleIDStrs := make([]string, len(in.WishedRoleIDs))
	for i, roleID := range in.WishedRoleIDs {
		wishedRoleIDStrs[i] = roleID.String()
	}

	eventPayload := outboxusecase.ParticipationRegisteredPayload{
		HackathonID:    in.HackathonID.String(),
		UserID:         userUUID.String(),
		Status:         in.DesiredStatus,
		WishedRoleIDs:  wishedRoleIDStrs,
		MotivationText: in.MotivationText,
		RegisteredAt:   now,
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event payload: %w", err)
	}

	outboxEvent := &outbox.Event{
		ID:            uuid.New(),
		AggregateID:   fmt.Sprintf("%s:%s", in.HackathonID.String(), userUUID.String()),
		AggregateType: "participation",
		EventType:     outboxusecase.EventTypeParticipationRegistered,
		Payload:       payloadBytes,
		Status:        outbox.EventStatusPending,
		AttemptCount:  0,
		LastError:     "",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := s.outboxRepo.Create(ctx, outboxEvent); err != nil {
		return nil, fmt.Errorf("failed to create outbox event: %w", err)
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
