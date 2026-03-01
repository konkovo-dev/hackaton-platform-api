package grpc

import (
	"context"
	"errors"
	"fmt"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	submissionv1 "github.com/belikoooova/hackaton-platform-api/api/submission/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/transport/grpc/mappers"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/usecase/submission"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SubmissionServiceServer struct {
	submissionv1.UnimplementedSubmissionServiceServer
	service *submission.Service
}

func NewSubmissionServiceServer(service *submission.Service) *SubmissionServiceServer {
	return &SubmissionServiceServer{
		service: service,
	}
}

func (s *SubmissionServiceServer) CreateSubmission(ctx context.Context, req *submissionv1.CreateSubmissionRequest) (*submissionv1.CreateSubmissionResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	out, err := s.service.CreateSubmission(ctx, submission.CreateSubmissionIn{
		HackathonID: hackathonID,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &submissionv1.CreateSubmissionResponse{
		SubmissionId: out.SubmissionID.String(),
		IsFinal:      out.IsFinal,
	}, nil
}

func (s *SubmissionServiceServer) UpdateSubmission(ctx context.Context, req *submissionv1.UpdateSubmissionRequest) (*submissionv1.UpdateSubmissionResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	submissionID, err := uuid.Parse(req.SubmissionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid submission_id")
	}

	out, err := s.service.UpdateSubmission(ctx, submission.UpdateSubmissionIn{
		HackathonID:  hackathonID,
		SubmissionID: submissionID,
		Description:  req.Description,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &submissionv1.UpdateSubmissionResponse{
		Submission: mappers.SubmissionToProto(out.Submission, nil),
	}, nil
}

func (s *SubmissionServiceServer) ListSubmissions(ctx context.Context, req *submissionv1.ListSubmissionsRequest) (*submissionv1.ListSubmissionsResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	var ownerKind string
	var ownerID uuid.UUID

	if req.OwnerKind != "" && req.OwnerId != "" {
		ownerKind = req.OwnerKind
		ownerID, err = uuid.Parse(req.OwnerId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
		}
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

	out, err := s.service.ListSubmissions(ctx, submission.ListSubmissionsIn{
		HackathonID: hackathonID,
		OwnerKind:   ownerKind,
		OwnerID:     ownerID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, mapError(err)
	}

	submissions := make([]*submissionv1.Submission, 0, len(out.Submissions))
	for _, sub := range out.Submissions {
		submissions = append(submissions, mappers.SubmissionToProto(sub, nil))
	}

	return &submissionv1.ListSubmissionsResponse{
		Submissions: submissions,
		Page: &commonv1.PageResponse{
			HasMore: out.HasMore,
		},
	}, nil
}

func (s *SubmissionServiceServer) SelectFinalSubmission(ctx context.Context, req *submissionv1.SelectFinalSubmissionRequest) (*submissionv1.SelectFinalSubmissionResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	submissionID, err := uuid.Parse(req.SubmissionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid submission_id")
	}

	out, err := s.service.SelectFinalSubmission(ctx, submission.SelectFinalSubmissionIn{
		HackathonID:  hackathonID,
		SubmissionID: submissionID,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &submissionv1.SelectFinalSubmissionResponse{
		SubmissionId: out.SubmissionID.String(),
	}, nil
}

func (s *SubmissionServiceServer) GetSubmission(ctx context.Context, req *submissionv1.GetSubmissionRequest) (*submissionv1.GetSubmissionResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	submissionID, err := uuid.Parse(req.SubmissionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid submission_id")
	}

	out, err := s.service.GetSubmission(ctx, submission.GetSubmissionIn{
		HackathonID:  hackathonID,
		SubmissionID: submissionID,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &submissionv1.GetSubmissionResponse{
		Submission: mappers.SubmissionToProto(out.Submission, out.Files),
	}, nil
}

func (s *SubmissionServiceServer) GetFinalSubmission(ctx context.Context, req *submissionv1.GetFinalSubmissionRequest) (*submissionv1.GetFinalSubmissionResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	ownerKind := req.OwnerKind
	ownerID, err := uuid.Parse(req.OwnerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
	}

	out, err := s.service.GetFinalSubmission(ctx, submission.GetFinalSubmissionIn{
		HackathonID: hackathonID,
		OwnerKind:   ownerKind,
		OwnerID:     ownerID,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &submissionv1.GetFinalSubmissionResponse{
		Submission: mappers.SubmissionToProto(out.Submission, out.Files),
	}, nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, submission.ErrUnauthorized):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, submission.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, submission.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, submission.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, submission.ErrConflict):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, submission.ErrTooManySubmissions):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, submission.ErrTooManyFiles):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, submission.ErrFileTooLarge):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, submission.ErrTotalSizeTooLarge):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, submission.ErrInvalidFileType):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, submission.ErrFileNotUploaded):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, fmt.Sprintf("internal error: %v", err))
	}
}
