package participationandrolesservice

import (
	"context"
	"errors"
	"log/slog"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/role"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *API) RejectStaffInvitation(ctx context.Context, req *participationrolesv1.RejectStaffInvitationRequest) (*participationrolesv1.RejectStaffInvitationResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.InvitationId)
		resp := &participationrolesv1.RejectStaffInvitationResponse{}
		found, err := a.idempotencyHelper.CheckAndGet(ctx, idempotencyKey, "reject_staff_invitation", requestHash, resp)
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

	invitationID, err := uuid.Parse(req.InvitationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid invitation_id")
	}

	_, err = a.roleService.RejectInvitation(ctx, role.RejectInvitationIn{
		InvitationID: invitationID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "reject_staff_invitation")
	}

	resp := &participationrolesv1.RejectStaffInvitationResponse{}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.InvitationId)
		if err := a.idempotencyHelper.Save(ctx, idempotencyKey, "reject_staff_invitation", requestHash, resp); err != nil {
			a.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	a.logger.InfoContext(ctx, "reject_staff_invitation: success",
		slog.String("invitation_id", req.InvitationId),
	)

	return resp, nil
}
