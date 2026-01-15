package skills

import "log/slog"

type Service struct {
	skillRepo SkillRepository
	logger    *slog.Logger
}

func NewService(skillRepo SkillRepository, logger *slog.Logger) *Service {
	return &Service{
		skillRepo: skillRepo,
		logger:    logger,
	}
}
