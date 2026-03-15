package judging

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain/entity"
	judgingpolicy "github.com/belikoooova/hackaton-platform-api/internal/judging-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type GetMyAssignmentsIn struct {
	HackathonID uuid.UUID
	Evaluated   *bool
	Limit       int32
	Offset      int32
}

type AssignmentWithSubmission struct {
	Assignment         *entity.Assignment
	IsEvaluated        bool
	SubmissionTitle    string
	SubmissionOwnerKind string
	SubmissionOwnerID  uuid.UUID
}

type GetMyAssignmentsOut struct {
	Assignments []*AssignmentWithSubmission
	TotalCount  int64
}

func (s *Service) GetMyAssignments(ctx context.Context, in GetMyAssignmentsIn) (*GetMyAssignmentsOut, error) {
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

	pctx, err := s.getMyAssignmentsPolicy.LoadContext(ctx, judgingpolicy.GetMyAssignmentsParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetActorRoles(roles)

	decision := s.getMyAssignmentsPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	var assignments []*entity.Assignment
	var isEvaluatedList []bool
	var totalCount int64

	if in.Evaluated != nil {
		assignments, isEvaluatedList, err = s.assignmentRepo.ListByJudgeFiltered(ctx, in.HackathonID, userUUID, *in.Evaluated, in.Limit, in.Offset)
		if err != nil {
			return nil, fmt.Errorf("failed to list assignments: %w", err)
		}

		totalCount, err = s.assignmentRepo.CountByJudgeFiltered(ctx, in.HackathonID, userUUID, *in.Evaluated)
		if err != nil {
			return nil, fmt.Errorf("failed to count assignments: %w", err)
		}
	} else {
		assignments, isEvaluatedList, err = s.assignmentRepo.ListByJudge(ctx, in.HackathonID, userUUID, in.Limit, in.Offset)
		if err != nil {
			return nil, fmt.Errorf("failed to list assignments: %w", err)
		}

		totalCount, err = s.assignmentRepo.CountByJudge(ctx, in.HackathonID, userUUID)
		if err != nil {
			return nil, fmt.Errorf("failed to count assignments: %w", err)
		}
	}

	result := make([]*AssignmentWithSubmission, 0, len(assignments))
	for i, assignment := range assignments {
		submission, err := s.submissionClient.GetSubmission(ctx, in.HackathonID.String(), assignment.SubmissionID.String())
		if err != nil {
			s.logger.Warn("failed to get submission for assignment",
				"assignment_id", assignment.ID.String(),
				"submission_id", assignment.SubmissionID.String(),
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

		result = append(result, &AssignmentWithSubmission{
			Assignment:          assignment,
			IsEvaluated:         isEvaluatedList[i],
			SubmissionTitle:     submission.Title,
			SubmissionOwnerKind: ownerKind,
			SubmissionOwnerID:   ownerUUID,
		})
	}

	return &GetMyAssignmentsOut{
		Assignments: result,
		TotalCount:  totalCount,
	}, nil
}
