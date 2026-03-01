package submission

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain/entity"
	submissionpolicy "github.com/belikoooova/hackaton-platform-api/internal/submission-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CreateSubmissionIn struct {
	HackathonID uuid.UUID
	Title       string
	Description string
}

type CreateSubmissionOut struct {
	SubmissionID uuid.UUID
	IsFinal      bool
}

func (s *Service) CreateSubmission(ctx context.Context, in CreateSubmissionIn) (*CreateSubmissionOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if in.Title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	stage, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	actorUserID, participationStatus, roles, teamID, err := s.prClient.GetHackathonContext(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon context: %w", err)
	}

	actorUUID, err := uuid.Parse(actorUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse actor user id: %w", err)
	}

	createPolicy := submissionpolicy.NewCreateSubmissionPolicy()
	pctx, err := createPolicy.LoadContext(ctx, submissionpolicy.CreateSubmissionParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(actorUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetParticipationStatus(participationStatus)
	pctx.SetActorRoles(roles)

	decision := createPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	ownerKind := domain.OwnerKindUser
	ownerID := userUUID

	if teamID != "" {
		ownerKind = domain.OwnerKindTeam
		teamUUID, err := uuid.Parse(teamID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse team id: %w", err)
		}
		ownerID = teamUUID
	}

	count, err := s.submissionRepo.CountByOwner(ctx, in.HackathonID, ownerKind, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to count submissions: %w", err)
	}

	if count >= int64(s.config.Limits.MaxSubmissionsPerOwner) {
		return nil, ErrTooManySubmissions
	}

	submissionID := uuid.New()
	submission := &entity.Submission{
		ID:              submissionID,
		HackathonID:     in.HackathonID,
		OwnerKind:       ownerKind,
		OwnerID:         ownerID,
		CreatedByUserID: userUUID,
		Title:           in.Title,
		Description:     in.Description,
		IsFinal:         true,
	}

	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		// Unset is_final flag from previous final submission
		if err := s.submissionRepo.UnsetFinalForOwner(ctx, in.HackathonID, ownerKind, ownerID); err != nil {
			return fmt.Errorf("failed to unset previous final submission: %w", err)
		}

		if err := s.submissionRepo.Create(ctx, submission); err != nil {
			return fmt.Errorf("failed to create submission: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.logger.Info("submission created",
		"submission_id", submissionID.String(),
		"hackathon_id", in.HackathonID.String(),
		"owner_kind", ownerKind,
		"owner_id", ownerID.String(),
		"is_final", true,
	)

	return &CreateSubmissionOut{
		SubmissionID: submissionID,
		IsFinal:      true,
	}, nil
}
