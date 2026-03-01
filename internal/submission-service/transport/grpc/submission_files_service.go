package grpc

import (
	"context"

	submissionv1 "github.com/belikoooova/hackaton-platform-api/api/submission/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/transport/grpc/mappers"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/usecase/submission"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SubmissionFilesServiceServer struct {
	submissionv1.UnimplementedSubmissionFilesServiceServer
	service *submission.Service
}

func NewSubmissionFilesServiceServer(service *submission.Service) *SubmissionFilesServiceServer {
	return &SubmissionFilesServiceServer{
		service: service,
	}
}

func (s *SubmissionFilesServiceServer) CreateSubmissionUpload(ctx context.Context, req *submissionv1.CreateSubmissionUploadRequest) (*submissionv1.CreateSubmissionUploadResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	submissionID, err := uuid.Parse(req.SubmissionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid submission_id")
	}

	out, err := s.service.CreateSubmissionUpload(ctx, submission.CreateSubmissionUploadIn{
		HackathonID:  hackathonID,
		SubmissionID: submissionID,
		Filename:     req.Filename,
		SizeBytes:    req.SizeBytes,
		ContentType:  req.ContentType,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &submissionv1.CreateSubmissionUploadResponse{
		FileId:    out.FileID.String(),
		UploadUrl: out.UploadURL,
		ExpiresAt: timestamppb.New(out.ExpiresAt),
	}, nil
}

func (s *SubmissionFilesServiceServer) CompleteSubmissionUpload(ctx context.Context, req *submissionv1.CompleteSubmissionUploadRequest) (*submissionv1.CompleteSubmissionUploadResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	submissionID, err := uuid.Parse(req.SubmissionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid submission_id")
	}

	fileID, err := uuid.Parse(req.FileId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid file_id")
	}

	out, err := s.service.CompleteSubmissionUpload(ctx, submission.CompleteSubmissionUploadIn{
		HackathonID:  hackathonID,
		SubmissionID: submissionID,
		FileID:       fileID,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &submissionv1.CompleteSubmissionUploadResponse{
		File: mappers.SubmissionFileToProto(out.File),
	}, nil
}

func (s *SubmissionFilesServiceServer) GetSubmissionFileDownloadURL(ctx context.Context, req *submissionv1.GetSubmissionFileDownloadURLRequest) (*submissionv1.GetSubmissionFileDownloadURLResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	submissionID, err := uuid.Parse(req.SubmissionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid submission_id")
	}

	fileID, err := uuid.Parse(req.FileId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid file_id")
	}

	out, err := s.service.GetSubmissionFileDownloadURL(ctx, submission.GetSubmissionFileDownloadURLIn{
		HackathonID:  hackathonID,
		SubmissionID: submissionID,
		FileID:       fileID,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &submissionv1.GetSubmissionFileDownloadURLResponse{
		DownloadUrl: out.DownloadURL,
		ExpiresAt:   timestamppb.New(out.ExpiresAt),
	}, nil
}
