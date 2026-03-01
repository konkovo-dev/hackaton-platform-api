package matchmaking

import (
	"context"
	"fmt"
	"sort"

	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type RecommendCandidatesRequest struct {
	UserID      uuid.UUID
	HackathonID uuid.UUID
	VacancyID   uuid.UUID
	Limit       int32
}

type RecommendCandidatesResponse struct {
	Recommendations []CandidateRecommendation
}

type CandidateRecommendation struct {
	Participation *entity.Participation
	User          *entity.User
	Score         entity.MatchScore
}

func (s *Service) RecommendCandidates(ctx context.Context, req RecommendCandidatesRequest) (*RecommendCandidatesResponse, error) {
	stage, err := s.hackathonClient.GetHackathon(ctx, req.HackathonID.String())
	if err != nil {
		if pgxutil.IsNotFound(err) {
			return nil, ErrHackathonNotFound
		}
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	if stage != domain.HackathonStageRegistration && stage != domain.HackathonStageRunning {
		return nil, ErrInvalidHackathonStage
	}

	_, teamIDPtr, err := s.participationClient.GetParticipationAndRoles(ctx, req.UserID.String(), req.HackathonID.String())
	if err != nil {
		if pgxutil.IsNotFound(err) {
			return nil, ErrParticipationNotFound
		}
		return nil, fmt.Errorf("failed to get participation and roles: %w", err)
	}

	if teamIDPtr == nil {
		return nil, ErrNotTeamCaptain
	}

	teamID, err := uuid.Parse(*teamIDPtr)
	if err != nil {
		return nil, fmt.Errorf("invalid team_id: %w", err)
	}

	captainUserID, err := s.teamClient.GetTeam(ctx, req.HackathonID.String(), teamID.String())
	if err != nil {
		if pgxutil.IsNotFound(err) {
			return nil, ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	if captainUserID != req.UserID.String() {
		return nil, ErrNotTeamCaptain
	}

	vacancy, err := s.vacancyRepo.GetByID(ctx, req.VacancyID)
	if err != nil {
		if pgxutil.IsNotFound(err) {
			return nil, ErrVacancyNotFound
		}
		return nil, fmt.Errorf("failed to get vacancy: %w", err)
	}

	if vacancy.TeamID != teamID {
		return nil, ErrVacancyNotFound
	}

	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		if pgxutil.IsNotFound(err) {
			return nil, ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	participations, err := s.participationRepo.ListByHackathon(ctx, req.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to list participations: %w", err)
	}

	userIDs := make([]uuid.UUID, 0, len(participations))
	for _, p := range participations {
		userIDs = append(userIDs, p.UserID)
	}

	users, err := s.userRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	usersByID := make(map[uuid.UUID]*entity.User)
	for _, user := range users {
		usersByID[user.UserID] = user
	}

	candidateScores := make([]CandidateRecommendation, 0)

	for _, participation := range participations {
		user, ok := usersByID[participation.UserID]
		if !ok {
			continue
		}

		score := s.calculateMatchScore(
			user.CatalogSkillIDs,
			participation.WishedRoleIDs,
			participation.MotivationText,
			user.CustomSkillNames,
			vacancy.DesiredSkillIDs,
			vacancy.DesiredRoleIDs,
			team.Description,
			vacancy.Description,
		)

		totalScore := score.Skills.Score*domain.ScoreWeightSkills +
			score.Roles.Score*domain.ScoreWeightRoles +
			score.Text.Score*domain.ScoreWeightText

		if totalScore > 0 {
			score.TotalScore = totalScore
			candidateScores = append(candidateScores, CandidateRecommendation{
				Participation: participation,
				User:          user,
				Score:         score,
			})
		}
	}

	sort.Slice(candidateScores, func(i, j int) bool {
		return candidateScores[i].Score.TotalScore > candidateScores[j].Score.TotalScore
	})

	if int32(len(candidateScores)) > req.Limit {
		candidateScores = candidateScores[:req.Limit]
	}

	return &RecommendCandidatesResponse{
		Recommendations: candidateScores,
	}, nil
}
