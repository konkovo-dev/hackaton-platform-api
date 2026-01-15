package skillsservice

import (
	"context"
	"errors"
	"log/slog"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/skills"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SkillsService struct {
	identityv1.UnimplementedSkillsServiceServer
	skillsService *skills.Service
	logger        *slog.Logger
}

var _ identityv1.SkillsServiceServer = (*SkillsService)(nil)

func NewSkillsService(
	skillsService *skills.Service,
	logger *slog.Logger,
) *SkillsService {
	return &SkillsService{
		skillsService: skillsService,
		logger:        logger,
	}
}

func (s *SkillsService) handleError(ctx context.Context, err error, operation string) error {
	switch {
	case errors.Is(err, skills.ErrInvalidInput):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		s.logger.ErrorContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Internal, "internal error")
	}
}
