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

type RecommendTeamsRequest struct {
	UserID      uuid.UUID
	HackathonID uuid.UUID
	Limit       int32
}

type RecommendTeamsResponse struct {
	Recommendations []TeamRecommendation
}

type TeamRecommendation struct {
	Team  *entity.Team
	Score entity.MatchScore
}

func (s *Service) RecommendTeams(ctx context.Context, req RecommendTeamsRequest) (*RecommendTeamsResponse, error) {
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

	participation, err := s.participationRepo.Get(ctx, req.HackathonID, req.UserID)
	if err != nil {
		if pgxutil.IsNotFound(err) {
			return nil, ErrParticipationNotFound
		}
		return nil, fmt.Errorf("failed to get participation: %w", err)
	}

	if participation.Status != domain.ParticipationStatusLookingForTeam {
		return nil, ErrNotLookingForTeam
	}

	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if pgxutil.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	teams, err := s.teamRepo.ListByHackathon(ctx, req.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}

	vacancies, err := s.vacancyRepo.ListByHackathon(ctx, req.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to list vacancies: %w", err)
	}

	vacanciesByTeam := make(map[uuid.UUID][]*entity.Vacancy)
	for _, vacancy := range vacancies {
		vacanciesByTeam[vacancy.TeamID] = append(vacanciesByTeam[vacancy.TeamID], vacancy)
	}

	teamScores := make([]TeamRecommendation, 0)

	for _, team := range teams {
		teamVacancies := vacanciesByTeam[team.TeamID]
		if len(teamVacancies) == 0 {
			continue
		}

		var bestScore entity.MatchScore
		var bestVacancyID uuid.UUID
		bestTotalScore := 0.0

		for _, vacancy := range teamVacancies {
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

			if totalScore > bestTotalScore {
				bestTotalScore = totalScore
				bestScore = score
				bestScore.TotalScore = totalScore
				bestVacancyID = vacancy.VacancyID
			}
		}

		if bestTotalScore > 0 {
			bestScore.BestVacancyID = &bestVacancyID
			teamScores = append(teamScores, TeamRecommendation{
				Team:  team,
				Score: bestScore,
			})
		}
	}

	sort.Slice(teamScores, func(i, j int) bool {
		return teamScores[i].Score.TotalScore > teamScores[j].Score.TotalScore
	})

	if int32(len(teamScores)) > req.Limit {
		teamScores = teamScores[:req.Limit]
	}

	return &RecommendTeamsResponse{
		Recommendations: teamScores,
	}, nil
}

func (s *Service) calculateMatchScore(
	userSkills []uuid.UUID,
	userRoles []uuid.UUID,
	userText string,
	userCustomSkills []string,
	desiredSkills []uuid.UUID,
	desiredRoles []uuid.UUID,
	teamDescription string,
	vacancyDescription string,
) entity.MatchScore {
	skillsBreakdown := s.calculateSkillsScore(userSkills, desiredSkills)
	rolesBreakdown := s.calculateRolesScore(userRoles, desiredRoles)
	textBreakdown := s.calculateTextScore(userText, userCustomSkills, teamDescription, vacancyDescription)

	return entity.MatchScore{
		Skills: skillsBreakdown,
		Roles:  rolesBreakdown,
		Text:   textBreakdown,
	}
}

func (s *Service) calculateSkillsScore(userSkills []uuid.UUID, desiredSkills []uuid.UUID) entity.SkillsBreakdown {
	if len(desiredSkills) == 0 {
		return entity.SkillsBreakdown{
			Score:         1.0,
			Weight:        domain.ScoreWeightSkills,
			MatchedSkills: []string{},
			MissingSkills: []string{},
			MatchedCount:  0,
			RequiredCount: 0,
		}
	}

	userSkillsSet := make(map[uuid.UUID]bool)
	for _, skill := range userSkills {
		userSkillsSet[skill] = true
	}

	matched := []string{}
	missing := []string{}

	for _, desiredSkill := range desiredSkills {
		if userSkillsSet[desiredSkill] {
			matched = append(matched, desiredSkill.String())
		} else {
			missing = append(missing, desiredSkill.String())
		}
	}

	matchRatio := 0.0
	if len(desiredSkills) > 0 {
		matchRatio = float64(len(matched)) / float64(len(desiredSkills))
	}

	return entity.SkillsBreakdown{
		Score:         matchRatio,
		Weight:        domain.ScoreWeightSkills,
		MatchedSkills: matched,
		MissingSkills: missing,
		MatchedCount:  int32(len(matched)),
		RequiredCount: int32(len(desiredSkills)),
	}
}

func (s *Service) calculateRolesScore(userRoles []uuid.UUID, desiredRoles []uuid.UUID) entity.RolesBreakdown {
	if len(desiredRoles) == 0 {
		return entity.RolesBreakdown{
			Score:         1.0,
			Weight:        domain.ScoreWeightRoles,
			MatchedRoles:  []string{},
			MatchedCount:  0,
			RequiredCount: 0,
		}
	}

	userRolesSet := make(map[uuid.UUID]bool)
	for _, role := range userRoles {
		userRolesSet[role] = true
	}

	matched := []string{}

	for _, desiredRole := range desiredRoles {
		if userRolesSet[desiredRole] {
			matched = append(matched, desiredRole.String())
		}
	}

	matchRatio := 0.0
	if len(desiredRoles) > 0 {
		matchRatio = float64(len(matched)) / float64(len(desiredRoles))
	}

	return entity.RolesBreakdown{
		Score:         matchRatio,
		Weight:        domain.ScoreWeightRoles,
		MatchedRoles:  matched,
		MatchedCount:  int32(len(matched)),
		RequiredCount: int32(len(desiredRoles)),
	}
}

func (s *Service) calculateTextScore(userText string, userCustomSkills []string, teamDescription string, vacancyDescription string) entity.TextBreakdown {
	if userText == "" && len(userCustomSkills) == 0 {
		return entity.TextBreakdown{
			Score:           0.0,
			Weight:          domain.ScoreWeightText,
			MatchedKeywords: []string{},
		}
	}

	userWords := extractWords(userText)
	for _, skill := range userCustomSkills {
		userWords = append(userWords, extractWords(skill)...)
	}

	if len(userWords) == 0 {
		return entity.TextBreakdown{
			Score:           0.0,
			Weight:          domain.ScoreWeightText,
			MatchedKeywords: []string{},
		}
	}

	teamWords := extractWords(teamDescription + " " + vacancyDescription)
	teamWordsSet := make(map[string]bool)
	for _, word := range teamWords {
		teamWordsSet[word] = true
	}

	matched := []string{}
	for _, word := range userWords {
		if teamWordsSet[word] {
			matched = append(matched, word)
		}
	}

	matchRatio := 0.0
	if len(userWords) > 0 {
		matchRatio = float64(len(matched)) / float64(len(userWords))
	}

	return entity.TextBreakdown{
		Score:           matchRatio,
		Weight:          domain.ScoreWeightText,
		MatchedKeywords: matched,
	}
}

func extractWords(text string) []string {
	words := make([]string, 0)
	current := ""

	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= 'а' && r <= 'я') || (r >= 'А' && r <= 'Я') ||
			(r >= '0' && r <= '9') {
			if r >= 'A' && r <= 'Z' {
				current += string(r + 32)
			} else if r >= 'А' && r <= 'Я' {
				current += string(r + 32)
			} else {
				current += string(r)
			}
		} else {
			if len(current) >= 3 {
				words = append(words, current)
			}
			current = ""
		}
	}

	if len(current) >= 3 {
		words = append(words, current)
	}

	return words
}
