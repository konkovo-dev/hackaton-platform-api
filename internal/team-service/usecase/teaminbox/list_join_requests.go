package teaminbox

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type ListJoinRequestsIn struct {
	HackathonID uuid.UUID
	TeamID      uuid.UUID
	PageSize    uint32
	PageToken   string
}

type ListJoinRequestsOut struct {
	JoinRequests  []*entity.JoinRequest
	NextPageToken string
}

func (s *Service) ListJoinRequests(ctx context.Context, in ListJoinRequestsIn) (*ListJoinRequestsOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	stage, _, _, err := s.hackathonClient.GetHackathon(ctx, in.HackathonID.String())
	if err != nil {
		s.logger.Error("failed to get hackathon", "error", err)
		return nil, fmt.Errorf("failed to get hackathon: %w", err)
	}

	_, err = s.teamRepo.GetByIDAndHackathonID(ctx, in.TeamID, in.HackathonID)
	if err != nil {
		return nil, ErrNotFound
	}

	isCaptain, err := s.membershipRepo.CheckIsCaptain(ctx, in.TeamID, userUUID)
	if err != nil {
		s.logger.Error("failed to check captain status", "error", err)
		return nil, fmt.Errorf("failed to check captain status: %w", err)
	}

	listPolicy := policy.NewListJoinRequestsPolicy()
	pctx, err := listPolicy.LoadContext(ctx, policy.ListJoinRequestsParams{
		HackathonID: in.HackathonID,
		TeamID:      in.TeamID,
	})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)
	pctx.SetHackathonStage(stage)
	pctx.SetIsCaptain(isCaptain)

	decision := listPolicy.Check(ctx, pctx)
	if !decision.Allowed {
		return nil, mapPolicyError(decision)
	}

	pageSize := in.PageSize
	if pageSize == 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	offset, err := parsePageToken(in.PageToken)
	if err != nil {
		return nil, fmt.Errorf("invalid page token: %w", err)
	}

	requests, err := s.joinRequestRepo.ListByTeamID(ctx, in.TeamID, int32(pageSize+1), int32(offset))
	if err != nil {
		s.logger.Error("failed to list join requests", "error", err)
		return nil, fmt.Errorf("failed to list join requests: %w", err)
	}

	var nextPageToken string
	if len(requests) > int(pageSize) {
		requests = requests[:pageSize]
		nextPageToken = encodePageToken(offset + int(pageSize))
	}

	return &ListJoinRequestsOut{
		JoinRequests:  requests,
		NextPageToken: nextPageToken,
	}, nil
}
