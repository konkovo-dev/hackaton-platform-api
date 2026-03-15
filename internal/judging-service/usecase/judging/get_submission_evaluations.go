package judging

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain/entity"
	judgingpolicy "github.com/belikoooova/hackaton-platform-api/internal/judging-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type GetSubmissionEvaluationsIn struct {
	HackathonID  uuid.UUID
	SubmissionID uuid.UUID
}

type GetSubmissionEvaluationsOut struct {
	Evaluations []*entity.Evaluation
}

func (s *Service) GetSubmissionEvaluations(ctx context.Context, in GetSubmissionEvaluationsIn) (*GetSubmissionEvaluationsOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	stage, _, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	_, _, roles, _, err := s.prClient.GetHackathonContext(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon context: %w", err)
	}

	pctx, err := s.getSubmissionEvalsPolicy.LoadContext(ctx, judgingpolicy.GetSubmissionEvaluationsParams{
		HackathonID:  in.HackathonID,
		SubmissionID: in.SubmissionID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetActorRoles(roles)

	decision := s.getSubmissionEvalsPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	evaluations, err := s.evaluationRepo.ListBySubmission(ctx, in.SubmissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list evaluations: %w", err)
	}

	return &GetSubmissionEvaluationsOut{
		Evaluations: evaluations,
	}, nil
}
