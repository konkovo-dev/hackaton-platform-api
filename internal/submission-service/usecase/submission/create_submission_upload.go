package submission

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain/entity"
	submissionpolicy "github.com/belikoooova/hackaton-platform-api/internal/submission-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type CreateSubmissionUploadIn struct {
	HackathonID  uuid.UUID
	SubmissionID uuid.UUID
	Filename     string
	SizeBytes    int64
	ContentType  string
}

type CreateSubmissionUploadOut struct {
	FileID    uuid.UUID
	UploadURL string
	ExpiresAt time.Time
}

func (s *Service) CreateSubmissionUpload(ctx context.Context, in CreateSubmissionUploadIn) (*CreateSubmissionUploadOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if in.Filename == "" {
		return nil, fmt.Errorf("%w: filename is required", ErrInvalidInput)
	}

	if in.SizeBytes <= 0 {
		return nil, fmt.Errorf("%w: size_bytes must be positive", ErrInvalidInput)
	}

	if !validateFileExtension(in.Filename, s.config.Limits.AllowedFileExtensions) {
		return nil, fmt.Errorf("%w: file extension not allowed", ErrInvalidFileType)
	}

	if !validateContentType(in.ContentType, s.config.Limits.AllowedContentTypes) {
		return nil, fmt.Errorf("%w: content type not allowed", ErrInvalidFileType)
	}

	if in.SizeBytes > s.config.Limits.MaxFileSizeBytes {
		return nil, fmt.Errorf("%w: max file size is %d bytes", ErrFileTooLarge, s.config.Limits.MaxFileSizeBytes)
	}

	submission, err := s.submissionRepo.GetByIDAndHackathonID(ctx, in.SubmissionID, in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}

	stage, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	_, _, _, teamID, err := s.prClient.GetHackathonContext(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon context: %w", err)
	}

	actorOwnerKind := domain.OwnerKindUser
	actorOwnerID := userUUID
	if teamID != "" {
		actorOwnerKind = domain.OwnerKindTeam
		teamUUID, err := uuid.Parse(teamID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse team id: %w", err)
		}
		actorOwnerID = teamUUID
	}

	fileOpPolicy := submissionpolicy.NewFileOperationPolicy(submissionpolicy.ActionCreateSubmissionUpload)
	pctx, err := fileOpPolicy.LoadContext(ctx, submissionpolicy.FileOperationParams{
		HackathonID:    in.HackathonID,
		SubmissionID:   in.SubmissionID,
		OwnerKind:      submission.OwnerKind,
		OwnerID:        submission.OwnerID,
		ActorOwnerKind: actorOwnerKind,
		ActorOwnerID:   actorOwnerID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)

	decision := fileOpPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	fileCount, err := s.fileRepo.CountBySubmission(ctx, in.SubmissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to count files: %w", err)
	}

	if fileCount >= int64(s.config.Limits.MaxFilesPerSubmission) {
		return nil, fmt.Errorf("%w: max %d files per submission", ErrTooManyFiles, s.config.Limits.MaxFilesPerSubmission)
	}

	totalSize, err := s.submissionRepo.GetTotalSize(ctx, in.SubmissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total size: %w", err)
	}

	if totalSize+in.SizeBytes > s.config.Limits.MaxTotalSizeBytes {
		return nil, fmt.Errorf("%w: max total size is %d bytes", ErrTotalSizeTooLarge, s.config.Limits.MaxTotalSizeBytes)
	}

	fileID := uuid.New()
	storageKey := fmt.Sprintf("%s/%s/%s/%s", in.HackathonID.String(), in.SubmissionID.String(), fileID.String(), in.Filename)

	file := &entity.SubmissionFile{
		ID:           fileID,
		SubmissionID: in.SubmissionID,
		Filename:     in.Filename,
		SizeBytes:    in.SizeBytes,
		ContentType:  in.ContentType,
		StorageKey:   storageKey,
		UploadStatus: domain.FileUploadStatusPending,
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	expiryDuration := time.Duration(s.config.Limits.PresignedURLExpiryMins) * time.Minute
	uploadURL, err := s.s3Client.GeneratePresignedPutURL(ctx, s.s3Client.Config().BucketName, storageKey, expiryDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	expiresAt := time.Now().UTC().Add(expiryDuration)

	s.logger.Info("file upload initiated",
		"file_id", fileID.String(),
		"submission_id", in.SubmissionID.String(),
		"filename", in.Filename,
		"size_bytes", in.SizeBytes,
	)

	return &CreateSubmissionUploadOut{
		FileID:    fileID,
		UploadURL: uploadURL,
		ExpiresAt: expiresAt,
	}, nil
}
