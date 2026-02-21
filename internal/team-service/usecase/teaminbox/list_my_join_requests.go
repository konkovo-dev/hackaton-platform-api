package teaminbox

import (
	"context"
	"fmt"

	"github.com/belikoooova/hackaton-platform-api/internal/team-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/policy"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
)

type ListMyJoinRequestsIn struct {
	PageSize  uint32
	PageToken string
}

type ListMyJoinRequestsOut struct {
	JoinRequests  []*entity.JoinRequest
	NextPageToken string
}

func (s *Service) ListMyJoinRequests(ctx context.Context, in ListMyJoinRequestsIn) (*ListMyJoinRequestsOut, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	listPolicy := policy.NewListMyJoinRequestsPolicy()
	pctx, err := listPolicy.LoadContext(ctx, policy.ListMyJoinRequestsParams{})
	if err != nil {
		return nil, err
	}

	pctx.SetAuthenticated(true)
	pctx.SetActorUserID(userUUID)

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

	requests, err := s.joinRequestRepo.ListByRequesterUserID(ctx, userUUID, int32(pageSize+1), int32(offset))
	if err != nil {
		s.logger.Error("failed to list my join requests", "error", err)
		return nil, fmt.Errorf("failed to list my join requests: %w", err)
	}

	var nextPageToken string
	if len(requests) > int(pageSize) {
		requests = requests[:pageSize]
		nextPageToken = encodePageToken(offset + int(pageSize))
	}

	return &ListMyJoinRequestsOut{
		JoinRequests:  requests,
		NextPageToken: nextPageToken,
	}, nil
}
