package judging

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/judging-service/domain"
	judgingpolicy "github.com/belikoooova/hackaton-platform-api/internal/judging-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type GetMyEvaluationResultIn struct {
	HackathonID uuid.UUID
}

type EvaluationResult struct {
	SubmissionID    uuid.UUID
	Title           string
	AverageScore    float64
	EvaluationCount int32
	Rank            int32
	Comments        []string
}

type GetMyEvaluationResultOut struct {
	Result *EvaluationResult
}

func (s *Service) GetMyEvaluationResult(ctx context.Context, in GetMyEvaluationResultIn) (*GetMyEvaluationResultOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	_, resultPublishedAt, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	_, participationStatus, _, teamID, err := s.prClient.GetHackathonContext(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon context: %w", err)
	}

	pctx, err := s.getMyResultPolicy.LoadContext(ctx, judgingpolicy.GetMyEvaluationResultParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetResultPublishedAt(resultPublishedAt != nil)
	pctx.SetParticipationStatus(participationStatus)

	decision := s.getMyResultPolicy.Check(ctx, pctx)
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

	// Get final submission for this owner from submission-service
	submissions, err := s.submissionClient.ListFinalSubmissions(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list final submissions: %w", err)
	}

	var finalSubmission *uuid.UUID
	var submissionTitle string
	for _, sub := range submissions {
		if ownerKindToString(sub.OwnerKind) == ownerKind && sub.OwnerId == ownerID.String() {
			subID, err := uuid.Parse(sub.SubmissionId)
			if err != nil {
				continue
			}
			finalSubmission = &subID
			submissionTitle = sub.Title
			break
		}
	}

	if finalSubmission == nil {
		return nil, ErrNotFound
	}

	// Get evaluations for this submission
	evaluations, err := s.leaderboardRepo.GetEvaluationsByOwner(ctx, in.HackathonID, *finalSubmission)
	if err != nil {
		return nil, fmt.Errorf("failed to get evaluations: %w", err)
	}

	if len(evaluations) == 0 {
		return nil, ErrNotFound
	}

	// Calculate average score and extract comments
	avgScore, evalCount, err := s.leaderboardRepo.GetSubmissionAverageScore(ctx, *finalSubmission)
	if err != nil {
		return nil, fmt.Errorf("failed to get average score: %w", err)
	}

	comments := make([]string, 0, len(evaluations))
	for _, eval := range evaluations {
		comments = append(comments, eval.Comment)
	}

	// Calculate rank by getting all scores
	allScores, err := s.leaderboardRepo.GetLeaderboardScores(ctx, in.HackathonID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get all scores for ranking: %w", err)
	}

	rank := int32(1)
	for _, score := range allScores {
		if score.SubmissionID == *finalSubmission {
			break
		}
		rank++
	}

	return &GetMyEvaluationResultOut{
		Result: &EvaluationResult{
			SubmissionID:    *finalSubmission,
			Title:           submissionTitle,
			AverageScore:    avgScore,
			EvaluationCount: evalCount,
			Rank:            rank,
			Comments:        comments,
		},
	}, nil
}
