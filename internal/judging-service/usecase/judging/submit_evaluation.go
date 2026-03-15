package judging

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain/entity"
	judgingpolicy "github.com/belikoooova/hackaton-platform-api/internal/judging-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type SubmitEvaluationIn struct {
	HackathonID  uuid.UUID
	SubmissionID uuid.UUID
	Score        int32
	Comment      string
}

type SubmitEvaluationOut struct {
	EvaluationID uuid.UUID
	EvaluatedAt  time.Time
}

func (s *Service) SubmitEvaluation(ctx context.Context, in SubmitEvaluationIn) (*SubmitEvaluationOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if in.Score < domain.MinScore || in.Score > domain.MaxScore {
		return nil, ErrInvalidScore
	}

	if len(in.Comment) < domain.MinCommentLength || len(in.Comment) > domain.MaxCommentLength {
		return nil, ErrInvalidComment
	}

	stage, _, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	_, _, roles, _, err := s.prClient.GetHackathonContext(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon context: %w", err)
	}

	isAssigned, err := s.assignmentRepo.CheckExists(ctx, in.HackathonID, in.SubmissionID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to check assignment: %w", err)
	}

	pctx, err := s.submitEvaluationPolicy.LoadContext(ctx, judgingpolicy.SubmitEvaluationParams{
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
	pctx.SetIsAssigned(isAssigned)

	decision := s.submitEvaluationPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	existingEval, err := s.evaluationRepo.GetBySubmissionAndJudge(ctx, in.SubmissionID, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing evaluation: %w", err)
	}

	if existingEval != nil {
		updatedEval, err := s.evaluationRepo.Update(ctx, existingEval.ID, in.Score, in.Comment)
		if err != nil {
			return nil, fmt.Errorf("failed to update evaluation: %w", err)
		}

		s.logger.Info("evaluation updated",
			"evaluation_id", updatedEval.ID.String(),
			"submission_id", in.SubmissionID.String(),
			"judge_user_id", userUUID.String(),
			"score", in.Score,
		)

		return &SubmitEvaluationOut{
			EvaluationID: updatedEval.ID,
			EvaluatedAt:  updatedEval.EvaluatedAt,
		}, nil
	}

	evaluationID := uuid.New()
	evaluation := &entity.Evaluation{
		ID:           evaluationID,
		HackathonID:  in.HackathonID,
		SubmissionID: in.SubmissionID,
		JudgeUserID:  userUUID,
		Score:        in.Score,
		Comment:      in.Comment,
	}

	if err := s.evaluationRepo.Create(ctx, evaluation); err != nil {
		return nil, fmt.Errorf("failed to create evaluation: %w", err)
	}

	s.logger.Info("evaluation created",
		"evaluation_id", evaluationID.String(),
		"hackathon_id", in.HackathonID.String(),
		"submission_id", in.SubmissionID.String(),
		"judge_user_id", userUUID.String(),
		"score", in.Score,
	)

	return &SubmitEvaluationOut{
		EvaluationID: evaluationID,
		EvaluatedAt:  evaluation.EvaluatedAt,
	}, nil
}
