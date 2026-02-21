package teammembersservice

import (
	"context"

	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/vacancy"
	"github.com/google/uuid"
)

func (a *API) UpsertVacancy(ctx context.Context, req *teamv1.UpsertVacancyRequest) (*teamv1.UpsertVacancyResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, a.handleError(ctx, err, "UpsertVacancy")
	}

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, a.handleError(ctx, err, "UpsertVacancy")
	}

	var vacancyID *uuid.UUID
	if req.VacancyId != "" {
		parsed, err := uuid.Parse(req.VacancyId)
		if err != nil {
			return nil, a.handleError(ctx, err, "UpsertVacancy")
		}
		vacancyID = &parsed
	}

	desiredRoleIDs := make([]uuid.UUID, 0, len(req.DesiredRoleIds))
	for _, roleID := range req.DesiredRoleIds {
		parsed, err := uuid.Parse(roleID)
		if err != nil {
			return nil, a.handleError(ctx, err, "UpsertVacancy")
		}
		desiredRoleIDs = append(desiredRoleIDs, parsed)
	}

	desiredSkillIDs := make([]uuid.UUID, 0, len(req.DesiredSkillIds))
	for _, skillID := range req.DesiredSkillIds {
		parsed, err := uuid.Parse(skillID)
		if err != nil {
			return nil, a.handleError(ctx, err, "UpsertVacancy")
		}
		desiredSkillIDs = append(desiredSkillIDs, parsed)
	}

	out, err := a.vacancyService.UpsertVacancy(ctx, vacancy.UpsertVacancyIn{
		HackathonID:     hackathonID,
		TeamID:          teamID,
		VacancyID:       vacancyID,
		Description:     req.Description,
		DesiredRoleIDs:  desiredRoleIDs,
		DesiredSkillIDs: desiredSkillIDs,
		SlotsTotal:      req.SlotsTotal,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "UpsertVacancy")
	}

	return &teamv1.UpsertVacancyResponse{
		VacancyId: out.VacancyID.String(),
	}, nil
}
