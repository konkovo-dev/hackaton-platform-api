package judging

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain/entity"
	judgingpolicy "github.com/belikoooova/hackaton-platform-api/internal/judging-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type GetMyEvaluationsIn struct {
	HackathonID uuid.UUID
	Limit       int32
	Offset      int32
}

type EvaluationWithSubmission struct {
	Evaluation          *entity.Evaluation
	SubmissionTitle     string
	SubmissionOwnerKind string
	SubmissionOwnerID   uuid.UUID
}

type GetMyEvaluationsOut struct {
	Evaluations []*EvaluationWithSubmission
	TotalCount  int64
}

func (s *Service) GetMyEvaluations(ctx context.Context, in GetMyEvaluationsIn) (*GetMyEvaluationsOut, error) {
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

	pctx, err := s.getMyEvaluationsPolicy.LoadContext(ctx, judgingpolicy.GetMyEvaluationsParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetActorRoles(roles)

	decision := s.getMyEvaluationsPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	evaluations, err := s.evaluationRepo.ListByJudge(ctx, in.HackathonID, userUUID, in.Limit, in.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list evaluations: %w", err)
	}

	totalCount, err := s.evaluationRepo.CountByJudge(ctx, in.HackathonID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to count evaluations: %w", err)
	}

	result := make([]*EvaluationWithSubmission, 0, len(evaluations))
	for _, evaluation := range evaluations {
		submission, err := s.submissionClient.GetSubmission(ctx, in.HackathonID.String(), evaluation.SubmissionID.String())
		if err != nil {
			s.logger.Warn("failed to get submission for evaluation",
				"evaluation_id", evaluation.ID.String(),
				"submission_id", evaluation.SubmissionID.String(),
				"error", err,
			)
			continue
		}

		ownerUUID, err := uuid.Parse(submission.OwnerId)
		if err != nil {
			s.logger.Warn("invalid owner id in submission",
				"submission_id", submission.SubmissionId,
				"owner_id", submission.OwnerId,
			)
			continue
		}

		ownerKind := "user"
		if submission.OwnerKind == 2 {
			ownerKind = "team"
		}

		result = append(result, &EvaluationWithSubmission{
			Evaluation:          evaluation,
			SubmissionTitle:     submission.Title,
			SubmissionOwnerKind: ownerKind,
			SubmissionOwnerID:   ownerUUID,
		})
	}

	return &GetMyEvaluationsOut{
		Evaluations: result,
		TotalCount:  totalCount,
	}, nil
}
