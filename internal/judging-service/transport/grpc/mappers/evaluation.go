package mappers

import (
	judgingv1 "github.com/belikoooova/hackaton-platform-api/api/judging/v1"
	submissionv1 "github.com/belikoooova/hackaton-platform-api/api/submission/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/usecase/judging"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func EvaluationToProto(e *entity.Evaluation) *judgingv1.Evaluation {
	return &judgingv1.Evaluation{
		EvaluationId: e.ID.String(),
		HackathonId:  e.HackathonID.String(),
		SubmissionId: e.SubmissionID.String(),
		JudgeUserId:  e.JudgeUserID.String(),
		Score:        e.Score,
		Comment:      e.Comment,
		EvaluatedAt:  timestamppb.New(e.EvaluatedAt),
		UpdatedAt:    timestamppb.New(e.UpdatedAt),
	}
}

func AssignmentToProto(a *entity.Assignment, isEvaluated bool) *judgingv1.Assignment {
	return &judgingv1.Assignment{
		AssignmentId: a.ID.String(),
		HackathonId:  a.HackathonID.String(),
		SubmissionId: a.SubmissionID.String(),
		JudgeUserId:  a.JudgeUserID.String(),
		IsEvaluated:  isEvaluated,
		AssignedAt:   timestamppb.New(a.AssignedAt),
	}
}

func AssignmentWithSubmissionToProto(a *judging.AssignmentWithSubmission) *judgingv1.AssignmentWithSubmission {
	ownerKind := submissionv1.OwnerKind_OWNER_KIND_USER
	if a.SubmissionOwnerKind == "team" {
		ownerKind = submissionv1.OwnerKind_OWNER_KIND_TEAM
	}

	return &judgingv1.AssignmentWithSubmission{
		Assignment:             AssignmentToProto(a.Assignment, a.IsEvaluated),
		SubmissionTitle:        a.SubmissionTitle,
		SubmissionOwnerKind:    ownerKind,
		SubmissionOwnerId:      a.SubmissionOwnerID.String(),
		SubmissionCreatedAt:    nil,
	}
}

func EvaluationWithSubmissionToProto(e *judging.EvaluationWithSubmission) *judgingv1.EvaluationWithSubmission {
	ownerKind := submissionv1.OwnerKind_OWNER_KIND_USER
	if e.SubmissionOwnerKind == "team" {
		ownerKind = submissionv1.OwnerKind_OWNER_KIND_TEAM
	}

	return &judgingv1.EvaluationWithSubmission{
		Evaluation:          EvaluationToProto(e.Evaluation),
		SubmissionTitle:     e.SubmissionTitle,
		SubmissionOwnerKind: ownerKind,
		SubmissionOwnerId:   e.SubmissionOwnerID.String(),
	}
}

func LeaderboardEntryToProto(e *judging.LeaderboardEntry) *judgingv1.LeaderboardEntry {
	ownerKind := submissionv1.OwnerKind_OWNER_KIND_USER
	if e.OwnerKind == "team" {
		ownerKind = submissionv1.OwnerKind_OWNER_KIND_TEAM
	}

	return &judgingv1.LeaderboardEntry{
		SubmissionId:    e.SubmissionID.String(),
		Title:           e.Title,
		OwnerKind:       ownerKind,
		OwnerId:         e.OwnerID.String(),
		AverageScore:    e.AverageScore,
		EvaluationCount: e.EvaluationCount,
		Rank:            e.Rank,
		CreatedAt:       nil,
	}
}

func EvaluationResultToProto(r *judging.EvaluationResult) *judgingv1.EvaluationResult {
	return &judgingv1.EvaluationResult{
		SubmissionId:    r.SubmissionID.String(),
		Title:           r.Title,
		AverageScore:    r.AverageScore,
		EvaluationCount: r.EvaluationCount,
		Rank:            r.Rank,
		Comments:        r.Comments,
	}
}
