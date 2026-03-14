package judging

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain/entity"
	judgingpolicy "github.com/belikoooova/hackaton-platform-api/internal/judging-service/policy"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/txrepo"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AssignSubmissionsToJudgesIn struct {
	HackathonID uuid.UUID
}

type AssignSubmissionsToJudgesOut struct {
	AssignmentsCount int32
	JudgesCount      int32
	SubmissionsCount int32
}

func (s *Service) AssignSubmissionsToJudges(ctx context.Context, in AssignSubmissionsToJudgesIn) (*AssignSubmissionsToJudgesOut, error) {
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

	pctx, err := s.assignPolicy.LoadContext(ctx, judgingpolicy.AssignSubmissionsParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetActorRoles(roles)

	decision := s.assignPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	existingCount, err := s.assignmentRepo.CountByHackathon(ctx, in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing assignments: %w", err)
	}

	if existingCount > 0 {
		s.logger.Info("assignments already exist, returning success",
			"hackathon_id", in.HackathonID.String(),
			"existing_count", existingCount,
		)
		return &AssignSubmissionsToJudgesOut{
			AssignmentsCount: int32(existingCount),
			JudgesCount:      0,
			SubmissionsCount: 0,
		}, nil
	}

	judgeIDs, err := s.prClient.ListJudges(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list judges: %w", err)
	}

	if len(judgeIDs) == 0 {
		return nil, ErrNoJudges
	}

	finalSubmissions, err := s.submissionClient.ListFinalSubmissions(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list final submissions: %w", err)
	}

	if len(finalSubmissions) == 0 {
		return nil, ErrNoSubmissions
	}

	judgesPerSubmission := domain.MinJudgesPerSubmission
	if len(judgeIDs) < judgesPerSubmission {
		judgesPerSubmission = len(judgeIDs)
	}

	var assignments []*entity.Assignment
	judgeIndex := 0

	for _, submission := range finalSubmissions {
		submissionUUID, err := uuid.Parse(submission.SubmissionId)
		if err != nil {
			s.logger.Warn("invalid submission id, skipping",
				"submission_id", submission.SubmissionId,
				"error", err,
			)
			continue
		}

		for i := 0; i < judgesPerSubmission; i++ {
			judgeUUID, err := uuid.Parse(judgeIDs[judgeIndex])
			if err != nil {
				s.logger.Warn("invalid judge id, skipping",
					"judge_id", judgeIDs[judgeIndex],
					"error", err,
				)
				judgeIndex = (judgeIndex + 1) % len(judgeIDs)
				continue
			}

			assignments = append(assignments, &entity.Assignment{
				ID:           uuid.New(),
				HackathonID:  in.HackathonID,
				SubmissionID: submissionUUID,
				JudgeUserID:  judgeUUID,
			})

			judgeIndex = (judgeIndex + 1) % len(judgeIDs)
		}
	}

	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		assignmentRepoTx := txrepo.NewAssignmentRepository(tx)
		for _, assignment := range assignments {
			if err := assignmentRepoTx.Create(ctx, assignment); err != nil {
				return fmt.Errorf("failed to create assignment: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.logger.Info("assignments created",
		"hackathon_id", in.HackathonID.String(),
		"assignments_count", len(assignments),
		"judges_count", len(judgeIDs),
		"submissions_count", len(finalSubmissions),
	)

	return &AssignSubmissionsToJudgesOut{
		AssignmentsCount: int32(len(assignments)),
		JudgesCount:      int32(len(judgeIDs)),
		SubmissionsCount: int32(len(finalSubmissions)),
	}, nil
}
