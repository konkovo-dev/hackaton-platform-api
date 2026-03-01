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

type CompleteSubmissionUploadIn struct {
	HackathonID  uuid.UUID
	SubmissionID uuid.UUID
	FileID       uuid.UUID
}

type CompleteSubmissionUploadOut struct {
	File *entity.SubmissionFile
}

func (s *Service) CompleteSubmissionUpload(ctx context.Context, in CompleteSubmissionUploadIn) (*CompleteSubmissionUploadOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	file, err := s.fileRepo.GetByIDAndSubmissionID(ctx, in.FileID, in.SubmissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	if file.UploadStatus != domain.FileUploadStatusPending {
		return nil, fmt.Errorf("%w: file upload is not in pending state", ErrConflict)
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

	fileOpPolicy := submissionpolicy.NewFileOperationPolicy(submissionpolicy.ActionCompleteSubmissionUpload)
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

	size, exists, err := s.s3Client.HeadObject(ctx, s.s3Client.Config().BucketName, file.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to verify file in S3: %w", err)
	}

	if !exists {
		completedAt := time.Now().UTC()
		updatedFile, err := s.fileRepo.UpdateStatus(ctx, in.FileID, domain.FileUploadStatusFailed, &completedAt)
		if err != nil {
			s.logger.Error("failed to mark file as failed", "file_id", in.FileID.String(), "error", err)
			return nil, fmt.Errorf("failed to update file status: %w", err)
		}
		s.logger.Warn("file not found in storage, marked as failed", "file_id", in.FileID.String(), "storage_key", file.StorageKey)
		return &CompleteSubmissionUploadOut{
			File: updatedFile,
		}, nil
	}

	if size != file.SizeBytes {
		completedAt := time.Now().UTC()
		updatedFile, err := s.fileRepo.UpdateStatus(ctx, in.FileID, domain.FileUploadStatusFailed, &completedAt)
		if err != nil {
			s.logger.Error("failed to mark file as failed", "file_id", in.FileID.String(), "error", err)
			return nil, fmt.Errorf("failed to update file status: %w", err)
		}
		s.logger.Warn("file size mismatch, marked as failed", "file_id", in.FileID.String(), "expected", file.SizeBytes, "actual", size)
		return &CompleteSubmissionUploadOut{
			File: updatedFile,
		}, nil
	}

	completedAt := time.Now().UTC()
	updatedFile, err := s.fileRepo.UpdateStatus(ctx, in.FileID, domain.FileUploadStatusCompleted, &completedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update file status: %w", err)
	}

	s.logger.Info("file upload completed",
		"file_id", in.FileID.String(),
		"submission_id", in.SubmissionID.String(),
		"filename", file.Filename,
	)

	return &CompleteSubmissionUploadOut{
		File: updatedFile,
	}, nil
}
