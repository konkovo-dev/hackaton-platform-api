package vacancy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	outboxusecase "github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/outbox"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	pkgpolicy "github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
)

type UpsertVacancyIn struct {
	HackathonID      uuid.UUID
	TeamID           uuid.UUID
	VacancyID        *uuid.UUID
	Description      string
	DesiredRoleIDs   []uuid.UUID
	DesiredSkillIDs  []uuid.UUID
	SlotsTotal       int64
}

type UpsertVacancyOut struct {
	VacancyID uuid.UUID
}

func (s *Service) UpsertVacancy(ctx context.Context, in UpsertVacancyIn) (*UpsertVacancyOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if in.SlotsTotal < 0 {
		return nil, fmt.Errorf("%w: slots_total must be non-negative", ErrBadRequest)
	}

	stage, allowTeam, teamSizeMax, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		s.logger.Error("failed to get hackathon", "error", err)
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	team, err := s.teamRepo.GetByIDAndHackathonID(ctx, in.TeamID, in.HackathonID)
	if err != nil {
		return nil, ErrNotFound
	}

	isCaptain, err := s.membershipRepo.CheckIsCaptain(ctx, in.TeamID, userUUID)
	if err != nil {
		s.logger.Error("failed to check captain status", "error", err)
		return nil, fmt.Errorf("failed to check captain status: %w", err)
	}

	upsertPolicy := policy.NewUpsertVacancyPolicy()
	pctx, err := upsertPolicy.LoadContext(ctx, policy.UpsertVacancyParams{
		HackathonID: in.HackathonID,
		TeamID:      in.TeamID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetAllowTeam(allowTeam)
	pctx.SetIsCaptain(isCaptain)

	decision := upsertPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	isUpdate := in.VacancyID != nil
	var vacancyID uuid.UUID
	var slotsOpen int64

	if isUpdate {
		vacancyID = *in.VacancyID
		oldVacancy, err := s.vacancyRepo.GetByID(ctx, vacancyID)
		if err != nil {
			return nil, ErrNotFound
		}

		if oldVacancy.TeamID != in.TeamID {
			return nil, ErrNotFound
		}

		occupiedSlots, err := s.vacancyRepo.CountOccupiedSlots(ctx, vacancyID)
		if err != nil {
			s.logger.Error("failed to count occupied slots", "error", err)
			return nil, fmt.Errorf("failed to count occupied slots: %w", err)
		}

		if in.SlotsTotal < occupiedSlots {
			return nil, fmt.Errorf("%w: slots_total cannot be less than currently occupied slots (%d)", ErrConflict, occupiedSlots)
		}

		slotsOpen = in.SlotsTotal - occupiedSlots
	} else {
		vacancyID = uuid.New()
		slotsOpen = in.SlotsTotal
	}

	membersCount, err := s.membershipRepo.CountMembers(ctx, in.TeamID)
	if err != nil {
		s.logger.Error("failed to count members", "error", err)
		return nil, fmt.Errorf("failed to count members: %w", err)
	}

	currentTotalOpenSlots, err := s.vacancyRepo.CountTotalOpenSlots(ctx, in.TeamID)
	if err != nil {
		s.logger.Error("failed to count total open slots", "error", err)
		return nil, fmt.Errorf("failed to count total open slots: %w", err)
	}

	var oldSlotsOpen int64
	if isUpdate {
		oldVacancy, _ := s.vacancyRepo.GetByID(ctx, vacancyID)
		oldSlotsOpen = oldVacancy.SlotsOpen
	}

	newTotalOpenSlots := currentTotalOpenSlots - oldSlotsOpen + slotsOpen
	totalCapacity := membersCount + newTotalOpenSlots

	if totalCapacity > int64(teamSizeMax) {
		return nil, fmt.Errorf("%w: team capacity (%d members + %d open slots = %d) exceeds hackathon limit (%d)",
			ErrConflict, membersCount, newTotalOpenSlots, totalCapacity, teamSizeMax)
	}

	now := time.Now().UTC()
	vacancy := &entity.Vacancy{
		ID:              vacancyID,
		TeamID:          team.ID,
		Description:     in.Description,
		DesiredRoleIDs:  in.DesiredRoleIDs,
		DesiredSkillIDs: in.DesiredSkillIDs,
		SlotsTotal:      in.SlotsTotal,
		SlotsOpen:       slotsOpen,
		IsSystem:        false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if isUpdate {
		err = s.vacancyRepo.Update(ctx, vacancy)
	} else {
		err = s.vacancyRepo.Create(ctx, vacancy)
	}

	if err != nil {
		s.logger.Error("failed to upsert vacancy", "error", err, "is_update", isUpdate)
		return nil, fmt.Errorf("failed to upsert vacancy: %w", err)
	}

	desiredRoleIDStrs := make([]string, len(in.DesiredRoleIDs))
	for i, roleID := range in.DesiredRoleIDs {
		desiredRoleIDStrs[i] = roleID.String()
	}

	desiredSkillIDStrs := make([]string, len(in.DesiredSkillIDs))
	for i, skillID := range in.DesiredSkillIDs {
		desiredSkillIDStrs[i] = skillID.String()
	}

	var eventType string
	var eventPayload interface{}

	if isUpdate {
		eventType = outboxusecase.EventTypeVacancyUpdated
		eventPayload = outboxusecase.VacancyUpdatedPayload{
			VacancyID:       vacancyID.String(),
			TeamID:          team.ID.String(),
			HackathonID:     team.HackathonID.String(),
			Description:     in.Description,
			DesiredRoleIDs:  desiredRoleIDStrs,
			DesiredSkillIDs: desiredSkillIDStrs,
			SlotsOpen:       int32(slotsOpen),
			UpdatedAt:       vacancy.UpdatedAt,
		}
	} else {
		eventType = outboxusecase.EventTypeVacancyCreated
		eventPayload = outboxusecase.VacancyCreatedPayload{
			VacancyID:       vacancyID.String(),
			TeamID:          team.ID.String(),
			HackathonID:     team.HackathonID.String(),
			Description:     in.Description,
			DesiredRoleIDs:  desiredRoleIDStrs,
			DesiredSkillIDs: desiredSkillIDStrs,
			SlotsOpen:       int32(slotsOpen),
			CreatedAt:       vacancy.CreatedAt,
		}
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event payload: %w", err)
	}

	outboxEvent := &outbox.Event{
		ID:            uuid.New(),
		AggregateID:   vacancyID.String(),
		AggregateType: "vacancy",
		EventType:     eventType,
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

	return &UpsertVacancyOut{VacancyID: vacancyID}, nil
}

func mapPolicyError(decision *pkgpolicy.Decision) error {
	if len(decision.Violations) == 0 {
		return ErrUnauthorized
	}

	v := decision.Violations[0]
	switch v.Code {
	case pkgpolicy.ViolationCodeForbidden:
		return fmt.Errorf("%w: %s", ErrForbidden, v.Message)
	case pkgpolicy.ViolationCodeNotFound:
		return fmt.Errorf("%w: %s", ErrNotFound, v.Message)
	case pkgpolicy.ViolationCodeStageRule, pkgpolicy.ViolationCodeTimeRule, pkgpolicy.ViolationCodePolicyRule:
		return fmt.Errorf("%w: %s", ErrForbidden, v.Message)
	case pkgpolicy.ViolationCodeConflict:
		return fmt.Errorf("%w: %s", ErrConflict, v.Message)
	default:
		return ErrUnauthorized
	}
}
