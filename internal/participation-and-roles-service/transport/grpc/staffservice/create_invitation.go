package staffservice

import (
	"context"
	"errors"
	"log/slog"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/role"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *API) CreateStaffInvitation(ctx context.Context, req *participationrolesv1.CreateStaffInvitationRequest) (*participationrolesv1.CreateStaffInvitationResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.TargetUserId, req.RequestedRole.String())
		resp := &participationrolesv1.CreateStaffInvitationResponse{}
		found, err := a.idempotencyHelper.CheckAndGet(ctx, idempotencyKey, "create_staff_invitation", requestHash, resp)
		if err != nil {
			var conflictErr *idempotency.ConflictError
			if errors.As(err, &conflictErr) {
				a.logger.WarnContext(ctx, "idempotency key conflict", slog.String("key", idempotencyKey))
				return nil, status.Error(codes.AlreadyExists, "idempotency key already used with different request")
			}
			a.logger.ErrorContext(ctx, "failed to check idempotency", slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, "failed to check idempotency")
		}
		if found {
			a.logger.InfoContext(ctx, "returning cached response", slog.String("key", idempotencyKey))
			return resp, nil
		}
	}

	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	targetUserID, err := uuid.Parse(req.TargetUserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid target_user_id")
	}

	protoRole := req.RequestedRole
	if protoRole == participationrolesv1.HackathonRole_HACKATHON_ROLE_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "invalid requested_role")
	}

	domainRole := domain.MapProtoRoleToDomain(protoRole)

	result, err := a.roleService.CreateInvitation(ctx, role.CreateInvitationIn{
		HackathonID:   hackathonID,
		TargetUserID:  targetUserID,
		RequestedRole: string(domainRole),
		Message:       req.Message,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "create_staff_invitation")
	}

	resp := &participationrolesv1.CreateStaffInvitationResponse{
		InvitationId: result.InvitationID.String(),
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.TargetUserId, req.RequestedRole.String())
		if err := a.idempotencyHelper.Save(ctx, idempotencyKey, "create_staff_invitation", requestHash, resp); err != nil {
			a.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	a.logger.InfoContext(ctx, "create_staff_invitation: success",
		slog.String("hackathon_id", req.HackathonId),
		slog.String("target_user_id", req.TargetUserId),
		slog.String("invitation_id", result.InvitationID.String()),
	)

	return resp, nil
}
