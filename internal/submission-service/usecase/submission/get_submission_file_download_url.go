package submission

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain"
	submissionpolicy "github.com/belikoooova/hackaton-platform-api/internal/submission-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type GetSubmissionFileDownloadURLIn struct {
	HackathonID  uuid.UUID
	SubmissionID uuid.UUID
	FileID       uuid.UUID
}

type GetSubmissionFileDownloadURLOut struct {
	DownloadURL string
	ExpiresAt   time.Time
}

func (s *Service) GetSubmissionFileDownloadURL(ctx context.Context, in GetSubmissionFileDownloadURLIn) (*GetSubmissionFileDownloadURLOut, error) {
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

	if file.UploadStatus != domain.FileUploadStatusCompleted {
		return nil, fmt.Errorf("%w: file upload is not completed", ErrConflict)
	}

	submission, err := s.submissionRepo.GetByIDAndHackathonID(ctx, in.SubmissionID, in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}

	_, _, roles, teamID, err := s.prClient.GetHackathonContext(ctx, in.HackathonID.String())
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

	downloadPolicy := submissionpolicy.NewGetDownloadURLPolicy()
	pctx, err := downloadPolicy.LoadContext(ctx, submissionpolicy.GetDownloadURLParams{
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
	pctx.SetActorRoles(roles)

	decision := downloadPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	expiryDuration := time.Duration(s.config.Limits.PresignedURLExpiryMins) * time.Minute
	downloadURL, err := s.s3Client.GeneratePresignedGetURL(ctx, s.s3Client.Config().BucketName, file.StorageKey, expiryDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	expiresAt := time.Now().UTC().Add(expiryDuration)

	s.logger.Info("file download URL generated",
		"file_id", in.FileID.String(),
		"submission_id", in.SubmissionID.String(),
	)

	return &GetSubmissionFileDownloadURLOut{
		DownloadURL: downloadURL,
		ExpiresAt:   expiresAt,
	}, nil
}
