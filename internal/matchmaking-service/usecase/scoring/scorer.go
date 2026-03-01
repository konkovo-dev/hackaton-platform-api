package scoring

import (
	"context"
	"fmt"
	"strings"

	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/repository/postgres/queries"
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
	"github.com/google/uuid"
)

type Scorer struct {
	*pgxutil.BaseRepository[*queries.Queries, queries.DBTX]
}

func NewScorer(db queries.DBTX) *Scorer {
	return &Scorer{
		BaseRepository: pgxutil.NewBaseRepository(db, queries.New),
	}
}

func (s *Scorer) ScoreTeamsForUser(
	ctx context.Context,
	userSkills []uuid.UUID,
	wishedRoles []uuid.UUID,
	motivationText string,
	userCustomSkills []string,
	teams []*entity.Team,
	vacanciesByTeam map[uuid.UUID][]*entity.Vacancy,
) ([]entity.TeamScore, error) {
	scores := make([]entity.TeamScore, 0, len(teams))

	for _, team := range teams {
		vacancies := vacanciesByTeam[team.TeamID]
		if len(vacancies) == 0 {
			continue
		}

		bestScore := entity.MatchScore{
			TotalScore: 0,
			Skills: entity.SkillsBreakdown{
				Score:  0,
				Weight: domain.ScoreWeightSkills,
			},
			Roles: entity.RolesBreakdown{
				Score:  0,
				Weight: domain.ScoreWeightRoles,
			},
			Text: entity.TextBreakdown{
				Score:  0,
				Weight: domain.ScoreWeightText,
			},
		}
		var bestVacancyID uuid.UUID

		for _, vacancy := range vacancies {
			score := s.calculateMatchScore(
				userSkills,
				wishedRoles,
				motivationText,
				userCustomSkills,
				vacancy.DesiredSkillIDs,
				vacancy.DesiredRoleIDs,
				team.Description,
				vacancy.Description,
			)

			if score.TotalScore > bestScore.TotalScore {
				bestScore = score
				bestVacancyID = vacancy.VacancyID
			}
		}

		if bestScore.TotalScore > 0 {
			bestScore.BestVacancyID = &bestVacancyID
			scores = append(scores, entity.TeamScore{
				Team:  team,
				Score: bestScore,
			})
		}
	}

	return scores, nil
}

func (s *Scorer) ScoreCandidatesForVacancy(
	ctx context.Context,
	vacancy *entity.Vacancy,
	team *entity.Team,
	participations []*entity.Participation,
	usersByID map[uuid.UUID]*entity.User,
) ([]entity.CandidateScore, error) {
	scores := make([]entity.CandidateScore, 0, len(participations))

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

		if score.TotalScore > 0 {
			scores = append(scores, entity.CandidateScore{
				Participation: participation,
				User:          user,
				Score:         score,
			})
		}
	}

	return scores, nil
}

func (s *Scorer) calculateMatchScore(
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

	totalScore := skillsBreakdown.Score*domain.ScoreWeightSkills +
		rolesBreakdown.Score*domain.ScoreWeightRoles +
		textBreakdown.Score*domain.ScoreWeightText

	return entity.MatchScore{
		TotalScore: totalScore,
		Skills:     skillsBreakdown,
		Roles:      rolesBreakdown,
		Text:       textBreakdown,
	}
}

func (s *Scorer) calculateSkillsScore(userSkills []uuid.UUID, desiredSkills []uuid.UUID) entity.SkillsBreakdown {
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

func (s *Scorer) calculateRolesScore(userRoles []uuid.UUID, desiredRoles []uuid.UUID) entity.RolesBreakdown {
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

func (s *Scorer) calculateTextScore(userText string, userCustomSkills []string, teamDescription string, vacancyDescription string) entity.TextBreakdown {
	userText = strings.ToLower(strings.TrimSpace(userText))
	teamDescription = strings.ToLower(strings.TrimSpace(teamDescription))
	vacancyDescription = strings.ToLower(strings.TrimSpace(vacancyDescription))

	if userText == "" && len(userCustomSkills) == 0 {
		return entity.TextBreakdown{
			Score:           0.0,
			Weight:          domain.ScoreWeightText,
			MatchedKeywords: []string{},
		}
	}

	userWords := extractWords(userText)
	for _, skill := range userCustomSkills {
		userWords = append(userWords, extractWords(strings.ToLower(skill))...)
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
	words := strings.Fields(text)
	result := make([]string, 0, len(words))

	for _, word := range words {
		cleaned := strings.Trim(word, ".,!?;:()[]{}\"'")
		if len(cleaned) >= 3 {
			result = append(result, cleaned)
		}
	}

	return result
}

type TeamScore struct {
	TeamID      uuid.UUID
	VacancyID   uuid.UUID
	TotalScore  float64
	SkillsScore float64
	RolesScore  float64
	TextScore   float64
}

func (s *Scorer) ScoreTeamsForUserSQL(
	ctx context.Context,
	userID uuid.UUID,
	hackathonID uuid.UUID,
	limit int32,
) ([]TeamScore, error) {
	rows, err := s.Queries().ScoreTeamsForUser(ctx, queries.ScoreTeamsForUserParams{
		UserID:      userID,
		HackathonID: hackathonID,
		Limit:       limit,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to score teams: %w", pgxutil.MapDBError(err))
	}

	scores := make([]TeamScore, 0, len(rows))
	for _, row := range rows {
		skillsScore, _ := row.SkillsScore.(float64)
		rolesScore, _ := row.RolesScore.(float64)
		textScore, _ := row.TextScore.(float64)
		
		scores = append(scores, TeamScore{
			TeamID:      row.TeamID,
			VacancyID:   row.VacancyID,
			TotalScore:  float64(row.TotalScore),
			SkillsScore: skillsScore,
			RolesScore:  rolesScore,
			TextScore:   textScore,
		})
	}

	return scores, nil
}

type CandidateScore struct {
	UserID      uuid.UUID
	TotalScore  float64
	SkillsScore float64
	RolesScore  float64
	TextScore   float64
}

func (s *Scorer) ScoreCandidatesForVacancySQL(
	ctx context.Context,
	vacancyID uuid.UUID,
	hackathonID uuid.UUID,
	limit int32,
) ([]CandidateScore, error) {
	rows, err := s.Queries().ScoreCandidatesForVacancy(ctx, queries.ScoreCandidatesForVacancyParams{
		VacancyID:   vacancyID,
		HackathonID: hackathonID,
		Limit:       limit,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to score candidates: %w", pgxutil.MapDBError(err))
	}

	scores := make([]CandidateScore, 0, len(rows))
	for _, row := range rows {
		skillsScore, _ := row.SkillsScore.(float64)
		rolesScore, _ := row.RolesScore.(float64)
		textScore, _ := row.TextScore.(float64)
		
		scores = append(scores, CandidateScore{
			UserID:      row.UserID,
			TotalScore:  float64(row.TotalScore),
			SkillsScore: skillsScore,
			RolesScore:  rolesScore,
			TextScore:   textScore,
		})
	}

	return scores, nil
}
