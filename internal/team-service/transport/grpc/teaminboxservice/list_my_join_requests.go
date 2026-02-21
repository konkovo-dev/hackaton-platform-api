package teaminboxservice

import (
	"context"

	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	teamv1 "github.com/belikoooova/hackaton-platform-api/api/team/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/team-service/usecase/teaminbox"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) ListMyJoinRequests(ctx context.Context, req *teamv1.ListMyJoinRequestsRequest) (*teamv1.ListMyJoinRequestsResponse, error) {
	var pageSize uint32
	var pageToken string
	if req.Query != nil && req.Query.Page != nil {
		pageSize = req.Query.Page.PageSize
		pageToken = req.Query.Page.PageToken
	}

	out, err := a.teamInboxService.ListMyJoinRequests(ctx, teaminbox.ListMyJoinRequestsIn{
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "ListMyJoinRequests")
	}

	protoRequests := make([]*teamv1.JoinRequest, 0, len(out.JoinRequests))
	for _, req := range out.JoinRequests {
		protoReq := &teamv1.JoinRequest{
			RequestId:       req.ID.String(),
			HackathonId:     req.HackathonID.String(),
			TeamId:          req.TeamID.String(),
			VacancyId:       req.VacancyID.String(),
			RequesterUserId: req.RequesterUserID.String(),
			Message:         req.Message,
			Status:          mapStatusToProto(req.Status),
			CreatedAt:       timestamppb.New(req.CreatedAt),
			UpdatedAt:       timestamppb.New(req.UpdatedAt),
		}

		if req.ExpiresAt != nil {
			protoReq.ExpiresAt = timestamppb.New(*req.ExpiresAt)
		}

		protoRequests = append(protoRequests, protoReq)
	}

	return &teamv1.ListMyJoinRequestsResponse{
		Requests: protoRequests,
		Page: &commonv1.PageResponse{
			NextPageToken: out.NextPageToken,
		},
	}, nil
}
