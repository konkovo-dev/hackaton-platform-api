package staffservice

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

func (a *API) CancelStaffInvitation(ctx context.Context, req *participationrolesv1.CancelStaffInvitationRequest) (*participationrolesv1.CancelStaffInvitationResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.InvitationId)
		resp := &participationrolesv1.CancelStaffInvitationResponse{}
		found, err := a.idempotencyHelper.CheckAndGet(ctx, idempotencyKey, "cancel_staff_invitation", requestHash, resp)
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

	invitationID, err := uuid.Parse(req.InvitationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid invitation_id")
	}

	_, err = a.roleService.CancelInvitation(ctx, role.CancelInvitationIn{
		HackathonID:  hackathonID,
		InvitationID: invitationID,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "cancel_staff_invitation")
	}

	resp := &participationrolesv1.CancelStaffInvitationResponse{}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.InvitationId)
		if err := a.idempotencyHelper.Save(ctx, idempotencyKey, "cancel_staff_invitation", requestHash, resp); err != nil {
			a.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	a.logger.InfoContext(ctx, "cancel_staff_invitation: success",
		slog.String("hackathon_id", req.HackathonId),
		slog.String("invitation_id", req.InvitationId),
	)

	return resp, nil
}
