package grpc

import (
	"context"
	"errors"
	"fmt"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	judgingv1 "github.com/belikoooova/hackaton-platform-api/api/judging/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/transport/grpc/mappers"
	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/usecase/judging"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type JudgingServiceServer struct {
	judgingv1.UnimplementedJudgingServiceServer
	service *judging.Service
}

func NewJudgingServiceServer(service *judging.Service) *JudgingServiceServer {
	return &JudgingServiceServer{
		service: service,
	}
}

func (s *JudgingServiceServer) AssignSubmissionsToJudges(ctx context.Context, req *judgingv1.AssignSubmissionsToJudgesRequest) (*judgingv1.AssignSubmissionsToJudgesResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	out, err := s.service.AssignSubmissionsToJudges(ctx, judging.AssignSubmissionsToJudgesIn{
		HackathonID: hackathonID,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &judgingv1.AssignSubmissionsToJudgesResponse{
		AssignmentsCount: out.AssignmentsCount,
		JudgesCount:      out.JudgesCount,
		SubmissionsCount: out.SubmissionsCount,
	}, nil
}

func (s *JudgingServiceServer) GetMyAssignments(ctx context.Context, req *judgingv1.GetMyAssignmentsRequest) (*judgingv1.GetMyAssignmentsResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	limit := int32(50)
	offset := int32(0)

	if req.Query != nil {
		if req.Query.Limit > 0 {
			limit = req.Query.Limit
		}
		if req.Query.Offset > 0 {
			offset = req.Query.Offset
		}
	}

	var evaluated *bool
	if req.Evaluated != nil {
		evaluated = req.Evaluated
	}

	out, err := s.service.GetMyAssignments(ctx, judging.GetMyAssignmentsIn{
		HackathonID: hackathonID,
		Evaluated:   evaluated,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, mapError(err)
	}

	assignments := make([]*judgingv1.AssignmentWithSubmission, 0, len(out.Assignments))
	for _, a := range out.Assignments {
		assignments = append(assignments, mappers.AssignmentWithSubmissionToProto(a))
	}

	hasMore := int64(offset+limit) < out.TotalCount

	return &judgingv1.GetMyAssignmentsResponse{
		Assignments: assignments,
		Page: &commonv1.PageResponse{
			HasMore: hasMore,
		},
	}, nil
}

func (s *JudgingServiceServer) SubmitEvaluation(ctx context.Context, req *judgingv1.SubmitEvaluationRequest) (*judgingv1.SubmitEvaluationResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	submissionID, err := uuid.Parse(req.SubmissionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid submission_id")
	}

	out, err := s.service.SubmitEvaluation(ctx, judging.SubmitEvaluationIn{
		HackathonID:  hackathonID,
		SubmissionID: submissionID,
		Score:        req.Score,
		Comment:      req.Comment,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &judgingv1.SubmitEvaluationResponse{
		EvaluationId: out.EvaluationID.String(),
		EvaluatedAt:  mappers.TimeToProto(out.EvaluatedAt),
	}, nil
}

func (s *JudgingServiceServer) GetMyEvaluations(ctx context.Context, req *judgingv1.GetMyEvaluationsRequest) (*judgingv1.GetMyEvaluationsResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	limit := int32(50)
	offset := int32(0)

	if req.Query != nil {
		if req.Query.Limit > 0 {
			limit = req.Query.Limit
		}
		if req.Query.Offset > 0 {
			offset = req.Query.Offset
		}
	}

	out, err := s.service.GetMyEvaluations(ctx, judging.GetMyEvaluationsIn{
		HackathonID: hackathonID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, mapError(err)
	}

	evaluations := make([]*judgingv1.EvaluationWithSubmission, 0, len(out.Evaluations))
	for _, e := range out.Evaluations {
		evaluations = append(evaluations, mappers.EvaluationWithSubmissionToProto(e))
	}

	hasMore := int64(offset+limit) < out.TotalCount

	return &judgingv1.GetMyEvaluationsResponse{
		Evaluations: evaluations,
		Page: &commonv1.PageResponse{
			HasMore: hasMore,
		},
	}, nil
}

func (s *JudgingServiceServer) GetSubmissionEvaluations(ctx context.Context, req *judgingv1.GetSubmissionEvaluationsRequest) (*judgingv1.GetSubmissionEvaluationsResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	submissionID, err := uuid.Parse(req.SubmissionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid submission_id")
	}

	out, err := s.service.GetSubmissionEvaluations(ctx, judging.GetSubmissionEvaluationsIn{
		HackathonID:  hackathonID,
		SubmissionID: submissionID,
	})
	if err != nil {
		return nil, mapError(err)
	}

	evaluations := make([]*judgingv1.Evaluation, 0, len(out.Evaluations))
	for _, e := range out.Evaluations {
		evaluations = append(evaluations, mappers.EvaluationToProto(e))
	}

	return &judgingv1.GetSubmissionEvaluationsResponse{
		Evaluations: evaluations,
	}, nil
}

func (s *JudgingServiceServer) GetLeaderboard(ctx context.Context, req *judgingv1.GetLeaderboardRequest) (*judgingv1.GetLeaderboardResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	limit := int32(50)
	offset := int32(0)

	if req.Query != nil {
		if req.Query.Limit > 0 {
			limit = req.Query.Limit
		}
		if req.Query.Offset > 0 {
			offset = req.Query.Offset
		}
	}

	out, err := s.service.GetLeaderboard(ctx, judging.GetLeaderboardIn{
		HackathonID: hackathonID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, mapError(err)
	}

	entries := make([]*judgingv1.LeaderboardEntry, 0, len(out.Entries))
	for _, e := range out.Entries {
		entries = append(entries, mappers.LeaderboardEntryToProto(e))
	}

	hasMore := int64(offset+limit) < out.TotalCount

	return &judgingv1.GetLeaderboardResponse{
		Entries: entries,
		Page: &commonv1.PageResponse{
			HasMore: hasMore,
		},
	}, nil
}

func (s *JudgingServiceServer) GetMyEvaluationResult(ctx context.Context, req *judgingv1.GetMyEvaluationResultRequest) (*judgingv1.GetMyEvaluationResultResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	out, err := s.service.GetMyEvaluationResult(ctx, judging.GetMyEvaluationResultIn{
		HackathonID: hackathonID,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &judgingv1.GetMyEvaluationResultResponse{
		Result: mappers.EvaluationResultToProto(out.Result),
	}, nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, judging.ErrUnauthorized):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, judging.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, judging.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, judging.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, judging.ErrConflict):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, judging.ErrNotAssigned):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, judging.ErrInvalidScore):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, judging.ErrInvalidComment):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, judging.ErrWrongStage):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, judging.ErrResultNotPublished):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, judging.ErrAlreadyAssigned):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, judging.ErrNoJudges):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, judging.ErrNoSubmissions):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, fmt.Sprintf("internal error: %v", err))
	}
}
