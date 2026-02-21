package participationservice

import (
	"context"
	"errors"
	"log/slog"

	participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/participation-and-roles-service/usecase/participation"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) RegisterForHackathon(ctx context.Context, req *participationrolesv1.RegisterForHackathonRequest) (*participationrolesv1.RegisterForHackathonResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.DesiredStatus.String())
		resp := &participationrolesv1.RegisterForHackathonResponse{}
		found, err := a.idempotencyHelper.CheckAndGet(ctx, idempotencyKey, "register_for_hackathon", requestHash, resp)
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

	protoStatus := req.DesiredStatus
	if protoStatus == participationrolesv1.ParticipationStatus_PARTICIPATION_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "desired_status is required")
	}

	domainStatus := mapProtoStatusToDomain(protoStatus)
	if domainStatus == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid desired_status")
	}

	wishedRoleIDs := make([]uuid.UUID, 0, len(req.WishedRoleIds))
	for _, roleIDStr := range req.WishedRoleIds {
		roleID, err := uuid.Parse(roleIDStr)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid wished_role_id")
		}
		wishedRoleIDs = append(wishedRoleIDs, roleID)
	}

	result, err := a.participationService.Register(ctx, participation.RegisterIn{
		HackathonID:    hackathonID,
		DesiredStatus:  domainStatus,
		WishedRoleIDs:  wishedRoleIDs,
		MotivationText: req.MotivationText,
	})
	if err != nil {
		return nil, a.handleError(ctx, err, "register_for_hackathon")
	}

	protoWishedRoles := make([]*participationrolesv1.TeamRole, 0, len(result.Participation.WishedRoles))
	for _, role := range result.Participation.WishedRoles {
		protoWishedRoles = append(protoWishedRoles, &participationrolesv1.TeamRole{
			Id:   role.ID.String(),
			Name: role.Name,
		})
	}

	resp := &participationrolesv1.RegisterForHackathonResponse{
		Participation: &participationrolesv1.HackathonParticipation{
			HackathonId: result.Participation.HackathonID.String(),
			UserId:      result.Participation.UserID.String(),
			Status:      mapDomainStatusToProto(result.Participation.Status),
			TeamId:      "",
			Profile: &participationrolesv1.ParticipationProfile{
				WishedRoles:    protoWishedRoles,
				MotivationText: result.Participation.MotivationText,
			},
			RegisteredAt: timestamppb.New(result.Participation.RegisteredAt),
			UpdatedAt:    timestamppb.New(result.Participation.UpdatedAt),
		},
	}

	if result.Participation.TeamID != nil {
		resp.Participation.TeamId = result.Participation.TeamID.String()
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.DesiredStatus.String())
		if err := a.idempotencyHelper.Save(ctx, idempotencyKey, "register_for_hackathon", requestHash, resp); err != nil {
			a.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	a.logger.InfoContext(ctx, "register_for_hackathon: success",
		slog.String("hackathon_id", req.HackathonId),
		slog.String("status", domainStatus),
	)

	return resp, nil
}

func mapProtoStatusToDomain(protoStatus participationrolesv1.ParticipationStatus) string {
	switch protoStatus {
	case participationrolesv1.ParticipationStatus_PART_NONE:
		return string(domain.ParticipationNone)
	case participationrolesv1.ParticipationStatus_PART_INDIVIDUAL:
		return string(domain.ParticipationIndividual)
	case participationrolesv1.ParticipationStatus_PART_LOOKING_FOR_TEAM:
		return string(domain.ParticipationLookingForTeam)
	case participationrolesv1.ParticipationStatus_PART_TEAM_MEMBER:
		return string(domain.ParticipationTeamMember)
	case participationrolesv1.ParticipationStatus_PART_TEAM_CAPTAIN:
		return string(domain.ParticipationTeamCaptain)
	default:
		return ""
	}
}

func mapDomainStatusToProto(domainStatus string) participationrolesv1.ParticipationStatus {
	switch domainStatus {
	case string(domain.ParticipationNone):
		return participationrolesv1.ParticipationStatus_PART_NONE
	case string(domain.ParticipationIndividual):
		return participationrolesv1.ParticipationStatus_PART_INDIVIDUAL
	case string(domain.ParticipationLookingForTeam):
		return participationrolesv1.ParticipationStatus_PART_LOOKING_FOR_TEAM
	case string(domain.ParticipationTeamMember):
		return participationrolesv1.ParticipationStatus_PART_TEAM_MEMBER
	case string(domain.ParticipationTeamCaptain):
		return participationrolesv1.ParticipationStatus_PART_TEAM_CAPTAIN
	default:
		return participationrolesv1.ParticipationStatus_PARTICIPATION_STATUS_UNSPECIFIED
	}
}
