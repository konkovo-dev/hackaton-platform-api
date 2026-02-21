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

func (a *API) SelfRemoveHackathonRole(ctx context.Context, req *participationrolesv1.SelfRemoveHackathonRoleRequest) (*participationrolesv1.SelfRemoveHackathonRoleResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.Role.String())
		resp := &participationrolesv1.SelfRemoveHackathonRoleResponse{}
		found, err := a.idempotencyHelper.CheckAndGet(ctx, idempotencyKey, "self_remove_hackathon_role", requestHash, resp)
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

	roleToRemove := domain.MapProtoRoleToDomain(req.Role)
	if roleToRemove == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	_, err = a.roleService.SelfRemoveRole(ctx, role.SelfRemoveRoleIn{
		HackathonID: hackathonID,
		Role:        string(roleToRemove),
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "self_remove_hackathon_role")
	}

	resp := &participationrolesv1.SelfRemoveHackathonRoleResponse{}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.Role.String())
		if err := a.idempotencyHelper.Save(ctx, idempotencyKey, "self_remove_hackathon_role", requestHash, resp); err != nil {
			a.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	a.logger.InfoContext(ctx, "self_remove_hackathon_role: success",
		slog.String("hackathon_id", req.HackathonId),
		slog.String("role", string(roleToRemove)),
	)

	return resp, nil
}
