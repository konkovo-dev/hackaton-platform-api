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

type UpdateSubmissionIn struct {
	HackathonID  uuid.UUID
	SubmissionID uuid.UUID
	Description  string
}

type UpdateSubmissionOut struct {
	Submission *entity.Submission
}

func (s *Service) UpdateSubmission(ctx context.Context, in UpdateSubmissionIn) (*UpdateSubmissionOut, error) {
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

	updatePolicy := submissionpolicy.NewUpdateSubmissionPolicy()
	pctx, err := updatePolicy.LoadContext(ctx, submissionpolicy.UpdateSubmissionParams{
		HackathonID:     in.HackathonID,
		SubmissionID:    in.SubmissionID,
		OwnerKind:       submission.OwnerKind,
		OwnerID:         submission.OwnerID,
		CreatedByUserID: submission.CreatedByUserID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetActorOwnerKind(actorOwnerKind)
	pctx.SetActorOwnerID(actorOwnerID)
	pctx.SetHackathonStage(stage)

	decision := updatePolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	updatedSubmission, err := s.submissionRepo.UpdateDescription(ctx, in.SubmissionID, in.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to update submission: %w", err)
	}

	s.logger.Info("submission updated",
		"submission_id", in.SubmissionID.String(),
		"hackathon_id", in.HackathonID.String(),
	)

	return &UpdateSubmissionOut{
		Submission: updatedSubmission,
	}, nil
}
