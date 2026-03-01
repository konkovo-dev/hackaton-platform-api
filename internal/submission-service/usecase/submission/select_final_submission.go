package submission

import (
	"context"
	"fmt"
	"time"

	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/domain"
	submissionpolicy "github.com/belikoooova/hackaton-platform-api/internal/submission-service/policy"
	"github.com/belikoooova/hackaton-platform-api/internal/submission-service/txrepo"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SelectFinalSubmissionIn struct {
	HackathonID  uuid.UUID
	SubmissionID uuid.UUID
}

type SelectFinalSubmissionOut struct {
	SubmissionID uuid.UUID
	SelectedAt   time.Time
}

func (s *Service) SelectFinalSubmission(ctx context.Context, in SelectFinalSubmissionIn) (*SelectFinalSubmissionOut, error) {
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

	var captainUserID *uuid.UUID
	if submission.OwnerKind == domain.OwnerKindTeam {
		captainIDStr, err := s.teamClient.GetTeamCaptain(ctx, in.HackathonID.String(), submission.OwnerID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get team captain: %w", err)
		}
		captainUUID, err := uuid.Parse(captainIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse captain id: %w", err)
		}
		captainUserID = &captainUUID
	}

	selectPolicy := submissionpolicy.NewSelectFinalSubmissionPolicy()
	pctx, err := selectPolicy.LoadContext(ctx, submissionpolicy.SelectFinalSubmissionParams{
		HackathonID:     in.HackathonID,
		SubmissionID:    in.SubmissionID,
		OwnerKind:       submission.OwnerKind,
		OwnerID:         submission.OwnerID,
		CreatedByUserID: submission.CreatedByUserID,
		CaptainUserID:   captainUserID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)

	decision := selectPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	selectedAt := time.Now().UTC()

	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		submissionRepoTx := txrepo.NewSubmissionRepository(tx)

		if err := submissionRepoTx.UnsetFinalSubmission(ctx, in.HackathonID, submission.OwnerKind, submission.OwnerID); err != nil {
			return fmt.Errorf("failed to unset previous final: %w", err)
		}

		if err := submissionRepoTx.SetFinalSubmission(ctx, in.SubmissionID); err != nil {
			return fmt.Errorf("failed to set final: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.logger.Info("final submission selected",
		"submission_id", in.SubmissionID.String(),
		"hackathon_id", in.HackathonID.String(),
		"owner_kind", submission.OwnerKind,
		"owner_id", submission.OwnerID.String(),
	)

	return &SelectFinalSubmissionOut{
		SubmissionID: in.SubmissionID,
		SelectedAt:   selectedAt,
	}, nil
}
