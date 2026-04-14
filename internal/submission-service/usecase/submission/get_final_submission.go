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

type GetFinalSubmissionIn struct {
	HackathonID uuid.UUID
	OwnerKind   string
	OwnerID     uuid.UUID
}

type GetFinalSubmissionOut struct {
	Submission *entity.Submission
	Files      []*entity.SubmissionFile
}

func (s *Service) GetFinalSubmission(ctx context.Context, in GetFinalSubmissionIn) (*GetFinalSubmissionOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	submission, err := s.submissionRepo.GetFinalByOwner(ctx, in.HackathonID, in.OwnerKind, in.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get final submission: %w", err)
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
		SubmissionID:   submission.ID,
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

	files, err := s.fileRepo.ListBySubmission(ctx, submission.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return &GetFinalSubmissionOut{
		Submission: submission,
		Files:      files,
	}, nil
}
