package judging

import (
	"context"
	"fmt"

	judgingpolicy "github.com/belikoooova/hackaton-platform-api/internal/judging-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type GetLeaderboardIn struct {
	HackathonID uuid.UUID
	Limit       int32
	Offset      int32
}

type LeaderboardEntry struct {
	SubmissionID    uuid.UUID
	Title           string
	OwnerKind       string
	OwnerID         uuid.UUID
	AverageScore    float64
	EvaluationCount int32
	Rank            int32
}

type GetLeaderboardOut struct {
	Entries    []*LeaderboardEntry
	TotalCount int64
}

func (s *Service) GetLeaderboard(ctx context.Context, in GetLeaderboardIn) (*GetLeaderboardOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	stage, _, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	_, _, roles, _, err := s.prClient.GetHackathonContext(ctx, in.HackathonID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get hackathon context: %w", err)
	}

	pctx, err := s.getLeaderboardPolicy.LoadContext(ctx, judgingpolicy.GetLeaderboardParams{
		HackathonID: in.HackathonID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetActorRoles(roles)

	decision := s.getLeaderboardPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	// Get scores from judging database
	s.logger.Info("getting leaderboard scores",
		"hackathon_id", in.HackathonID.String(),
		"limit", in.Limit,
		"offset", in.Offset,
	)
	scoreEntries, err := s.leaderboardRepo.GetLeaderboardScores(ctx, in.HackathonID, in.Limit, in.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard scores: %w", err)
	}
	s.logger.Info("got leaderboard scores",
		"hackathon_id", in.HackathonID.String(),
		"count", len(scoreEntries),
	)
	for i, entry := range scoreEntries {
		s.logger.Info("leaderboard entry",
			"index", i,
			"submission_id", entry.SubmissionID.String(),
			"average_score", entry.AverageScore,
			"evaluation_count", entry.EvaluationCount,
		)
	}

	totalCount, err := s.leaderboardRepo.CountEvaluatedSubmissions(ctx, in.HackathonID)
	if err != nil {
		return nil, fmt.Errorf("failed to count evaluated submissions: %w", err)
	}

	// Get submission details from submission-service
	result := make([]*LeaderboardEntry, 0, len(scoreEntries))
	for i, scoreEntry := range scoreEntries {
		submission, err := s.submissionClient.GetSubmission(ctx, in.HackathonID.String(), scoreEntry.SubmissionID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get submission %s: %w", scoreEntry.SubmissionID, err)
		}

		rank := int32(in.Offset) + int32(i) + 1

		result = append(result, &LeaderboardEntry{
			SubmissionID:    scoreEntry.SubmissionID,
			Title:           submission.Title,
			OwnerKind:       ownerKindToString(submission.OwnerKind),
			OwnerID:         uuid.MustParse(submission.OwnerId),
			AverageScore:    scoreEntry.AverageScore,
			EvaluationCount: scoreEntry.EvaluationCount,
			Rank:            rank,
		})
	}

	return &GetLeaderboardOut{
		Entries:    result,
		TotalCount: totalCount,
	}, nil
}
