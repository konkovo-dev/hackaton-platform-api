package mappers

import (
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain/entity"
	submissionv1 "github.com/belikoooova/hackaton-platform-api/api/submission/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func SubmissionToProto(s *entity.Submission, files []*entity.SubmissionFile) *submissionv1.Submission {
	protoFiles := make([]*submissionv1.SubmissionFile, 0, len(files))
	for _, f := range files {
		protoFiles = append(protoFiles, SubmissionFileToProto(f))
	}

	ownerKind := submissionv1.OwnerKind_OWNER_KIND_UNSPECIFIED
	switch s.OwnerKind {
	case "user":
		ownerKind = submissionv1.OwnerKind_OWNER_KIND_USER
	case "team":
		ownerKind = submissionv1.OwnerKind_OWNER_KIND_TEAM
	}

	return &submissionv1.Submission{
		SubmissionId:    s.ID.String(),
		HackathonId:     s.HackathonID.String(),
		OwnerKind:       ownerKind,
		OwnerId:         s.OwnerID.String(),
		CreatedByUserId: s.CreatedByUserID.String(),
		Title:           s.Title,
		Description:     s.Description,
		IsFinal:         s.IsFinal,
		Files:           protoFiles,
		CreatedAt:       timestamppb.New(s.CreatedAt),
		UpdatedAt:       timestamppb.New(s.UpdatedAt),
	}
}

func SubmissionFileToProto(f *entity.SubmissionFile) *submissionv1.SubmissionFile {
	uploadStatus := submissionv1.FileUploadStatus_FILE_UPLOAD_STATUS_UNSPECIFIED
	switch f.UploadStatus {
	case "pending":
		uploadStatus = submissionv1.FileUploadStatus_FILE_UPLOAD_STATUS_PENDING
	case "completed":
		uploadStatus = submissionv1.FileUploadStatus_FILE_UPLOAD_STATUS_COMPLETED
	case "failed":
		uploadStatus = submissionv1.FileUploadStatus_FILE_UPLOAD_STATUS_FAILED
	}

	var completedAt *timestamppb.Timestamp
	if f.CompletedAt != nil {
		completedAt = timestamppb.New(*f.CompletedAt)
	}

	return &submissionv1.SubmissionFile{
		FileId:       f.ID.String(),
		SubmissionId: f.SubmissionID.String(),
		Filename:     f.Filename,
		SizeBytes:    f.SizeBytes,
		ContentType:  f.ContentType,
		UploadStatus: uploadStatus,
		CreatedAt:    timestamppb.New(f.CreatedAt),
		CompletedAt:  completedAt,
	}
}
