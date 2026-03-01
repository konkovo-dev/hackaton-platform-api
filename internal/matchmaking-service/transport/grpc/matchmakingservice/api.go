package matchmakingservice

import (
	"context"
	"errors"
	"log/slog"

	matchmakingv1 "github.com/belikoooova/hackaton-platform-api/api/matchmaking/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/matchmaking-service/usecase/matchmaking"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type API struct {
	matchmakingv1.UnimplementedMatchmakingServiceServer
	service *matchmaking.Service
	logger  *slog.Logger
}

var _ matchmakingv1.MatchmakingServiceServer = (*API)(nil)

func NewAPI(service *matchmaking.Service, logger *slog.Logger) *API {
	return &API{
		service: service,
		logger:  logger,
	}
}

func (a *API) RecommendTeams(ctx context.Context, req *matchmakingv1.RecommendTeamsRequest) (*matchmakingv1.RecommendTeamsResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	userIDStr, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	limit := int32(20)
	if req.Limit > 0 {
		limit = req.Limit
	}

	output, err := a.service.RecommendTeams(ctx, matchmaking.RecommendTeamsRequest{
		UserID:      userID,
		HackathonID: hackathonID,
		Limit:       limit,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "RecommendTeams")
	}

	recommendations := make([]*matchmakingv1.TeamRecommendation, 0, len(output.Recommendations))
	for _, rec := range output.Recommendations {
		bestVacancyID := ""
		if rec.Score.BestVacancyID != nil {
			bestVacancyID = rec.Score.BestVacancyID.String()
		}

		recommendations = append(recommendations, &matchmakingv1.TeamRecommendation{
			TeamId:        rec.Team.TeamID.String(),
			BestVacancyId: bestVacancyID,
			MatchScore:    mapMatchScoreToProto(&rec.Score),
		})
	}

	return &matchmakingv1.RecommendTeamsResponse{
		Recommendations: recommendations,
	}, nil
}

func (a *API) RecommendCandidates(ctx context.Context, req *matchmakingv1.RecommendCandidatesRequest) (*matchmakingv1.RecommendCandidatesResponse, error) {
	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	userIDStr, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	vacancyID, err := uuid.Parse(req.VacancyId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vacancy_id")
	}

	limit := int32(20)
	if req.Limit > 0 {
		limit = req.Limit
	}

	output, err := a.service.RecommendCandidates(ctx, matchmaking.RecommendCandidatesRequest{
		UserID:      userID,
		HackathonID: hackathonID,
		VacancyID:   vacancyID,
		Limit:       limit,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "RecommendCandidates")
	}

	recommendations := make([]*matchmakingv1.CandidateRecommendation, 0, len(output.Recommendations))
	for _, rec := range output.Recommendations {
		recommendations = append(recommendations, &matchmakingv1.CandidateRecommendation{
			UserId:     rec.User.UserID.String(),
			MatchScore: mapMatchScoreToProto(&rec.Score),
		})
	}

	return &matchmakingv1.RecommendCandidatesResponse{
		Recommendations: recommendations,
	}, nil
}

func mapMatchScoreToProto(score *entity.MatchScore) *matchmakingv1.MatchScore {
	protoScore := &matchmakingv1.MatchScore{
		TotalScore: score.TotalScore,
		Skills: &matchmakingv1.SkillsBreakdown{
			Score:         score.Skills.Score,
			Weight:        score.Skills.Weight,
			MatchedSkills: score.Skills.MatchedSkills,
			MissingSkills: score.Skills.MissingSkills,
			MatchedCount:  score.Skills.MatchedCount,
			RequiredCount: score.Skills.RequiredCount,
		},
		Roles: &matchmakingv1.RolesBreakdown{
			Score:         score.Roles.Score,
			Weight:        score.Roles.Weight,
			MatchedRoles:  score.Roles.MatchedRoles,
			MatchedCount:  score.Roles.MatchedCount,
			RequiredCount: score.Roles.RequiredCount,
		},
		Text: &matchmakingv1.TextBreakdown{
			Score:           score.Text.Score,
			Weight:          score.Text.Weight,
			MatchedKeywords: score.Text.MatchedKeywords,
		},
	}

	if score.BestVacancyID != nil {
		protoScore.BestVacancyId = score.BestVacancyID.String()
	}

	return protoScore
}

func (a *API) handleError(ctx context.Context, err error, method string) error {
	a.logger.ErrorContext(ctx, "error in "+method, "error", err)

	if errors.Is(err, matchmaking.ErrHackathonNotFound) {
		return status.Error(codes.NotFound, "hackathon not found")
	}
	if errors.Is(err, matchmaking.ErrParticipationNotFound) {
		return status.Error(codes.NotFound, "participation not found")
	}
	if errors.Is(err, matchmaking.ErrTeamNotFound) {
		return status.Error(codes.NotFound, "team not found")
	}
	if errors.Is(err, matchmaking.ErrVacancyNotFound) {
		return status.Error(codes.NotFound, "vacancy not found")
	}
	if errors.Is(err, matchmaking.ErrUserNotFound) {
		return status.Error(codes.NotFound, "user not found")
	}
	if errors.Is(err, matchmaking.ErrInvalidHackathonStage) {
		return status.Error(codes.FailedPrecondition, "matchmaking is not available at this hackathon stage")
	}
	if errors.Is(err, matchmaking.ErrNotLookingForTeam) {
		return status.Error(codes.FailedPrecondition, "user is not looking for team")
	}
	if errors.Is(err, matchmaking.ErrNotTeamCaptain) {
		return status.Error(codes.PermissionDenied, "user is not team captain")
	}

	var policyErr *policy.PolicyError
	if errors.As(err, &policyErr) {
		return status.Error(codes.PermissionDenied, policyErr.Error())
	}

	return status.Error(codes.Internal, "internal server error")
}
