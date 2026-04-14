package submission

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain/entity"
	submissionpolicy "github.com/belikoooova/hackaton-platform-api/internal/submission-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type GetSubmissionIn struct {
	HackathonID  uuid.UUID
	SubmissionID uuid.UUID
}

type GetSubmissionOut struct {
	Submission *entity.Submission
	Files      []*entity.SubmissionFile
}

func (s *Service) GetSubmission(ctx context.Context, in GetSubmissionIn) (*GetSubmissionOut, error) {
	// Service-to-service calls bypass policy checks
	if auth.IsServiceCall(ctx) {
		submission, err := s.submissionRepo.GetByIDAndHackathonID(ctx, in.SubmissionID, in.HackathonID)
		if err != nil {
			return nil, fmt.Errorf("failed to get submission: %w", err)
		}

		files, err := s.fileRepo.ListBySubmission(ctx, in.SubmissionID)
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		return &GetSubmissionOut{
			Submission: submission,
			Files:      files,
		}, nil
	}

	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	submission, err := s.submissionRepo.GetByIDAndHackathonID(ctx, in.SubmissionID, in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}

	stage, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon stage: %w", err)
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

	getPolicy := submissionpolicy.NewGetSubmissionPolicy()
	pctx, err := getPolicy.LoadContext(ctx, submissionpolicy.GetSubmissionParams{
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
	pctx.SetHackathonStage(stage)

	decision := getPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	files, err := s.fileRepo.ListBySubmission(ctx, in.SubmissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return &GetSubmissionOut{
		Submission: submission,
		Files:      files,
	}, nil
}
