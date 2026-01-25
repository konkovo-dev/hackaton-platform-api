package participationandrolesservice

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

func (a *API) AssignHackathonRole(ctx context.Context, req *participationrolesv1.AssignHackathonRoleRequest) (*participationrolesv1.AssignHackathonRoleResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.UserId, req.Role.String())
		resp := &participationrolesv1.AssignHackathonRoleResponse{}
		found, err := a.idempotencyHelper.CheckAndGet(ctx, idempotencyKey, "assign_hackathon_role", requestHash, resp)
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

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	roleStr := string(domain.MapProtoRoleToDomain(req.Role))
	if roleStr == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	err = a.roleService.AssignRole(ctx, role.AssignRoleIn{
		HackathonID: hackathonID,
		UserID:      userID,
		Role:        roleStr,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "assign_hackathon_role")
	}

	resp := &participationrolesv1.AssignHackathonRoleResponse{}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.UserId, req.Role.String())
		if err := a.idempotencyHelper.Save(ctx, idempotencyKey, "assign_hackathon_role", requestHash, resp); err != nil {
			a.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	a.logger.InfoContext(ctx, "hackathon role assigned",
		slog.String("hackathon_id", req.HackathonId),
		slog.String("user_id", req.UserId),
		slog.String("role", roleStr),
	)
	return resp, nil
}
