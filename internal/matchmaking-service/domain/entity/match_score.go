package entity

import "github.com/google/uuid"

type MatchScore struct {
	TotalScore     float64
	Skills         SkillsBreakdown
	Roles          RolesBreakdown
	Text           TextBreakdown
	BestVacancyID  *uuid.UUID
}

type SkillsBreakdown struct {
	Score          float64
	Weight         float64
	MatchedSkills  []string
	MissingSkills  []string
	MatchedCount   int32
	RequiredCount  int32
}

type RolesBreakdown struct {
	Score         float64
	Weight        float64
	MatchedRoles  []string
	MatchedCount  int32
	RequiredCount int32
}

type TextBreakdown struct {
	Score           float64
	Weight          float64
	MatchedKeywords []string
}
