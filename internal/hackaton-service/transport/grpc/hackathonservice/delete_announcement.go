package hackathonservice

import (
	"context"
	"errors"
	"log/slog"

	hackathonv1 "github.com/belikoooova/hackaton-platform-api/api/hackathon/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/hackaton-service/usecase/announcement"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *HackathonService) DeleteHackathonAnnouncement(ctx context.Context, req *hackathonv1.DeleteHackathonAnnouncementRequest) (*hackathonv1.DeleteHackathonAnnouncementResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.AnnouncementId, "delete")
		resp := &hackathonv1.DeleteHackathonAnnouncementResponse{}
		found, err := s.idempotency.CheckAndGet(ctx, idempotencyKey, "delete_announcement", requestHash, resp)
		if err != nil {
			var conflictErr *idempotency.ConflictError
			if errors.As(err, &conflictErr) {
				s.logger.WarnContext(ctx, "idempotency key conflict", slog.String("key", idempotencyKey))
				return nil, status.Error(codes.AlreadyExists, "idempotency key already used with different request")
			}
			s.logger.ErrorContext(ctx, "failed to check idempotency", slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, "failed to check idempotency")
		}
		if found {
			s.logger.InfoContext(ctx, "returning cached response", slog.String("key", idempotencyKey))
			return resp, nil
		}
	}

	hackathonID, err := uuid.Parse(req.HackathonId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hackathon_id")
	}

	announcementID, err := uuid.Parse(req.AnnouncementId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid announcement_id")
	}

	_, err = s.announcementService.DeleteAnnouncement(ctx, announcement.DeleteAnnouncementIn{
		HackathonID:    hackathonID,
		AnnouncementID: announcementID,
	})
	if err != nil {
		return nil, s.handleAnnouncementError(ctx, err, "delete_announcement")
	}

	resp := &hackathonv1.DeleteHackathonAnnouncementResponse{}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.AnnouncementId, "delete")
		if err := s.idempotency.Save(ctx, idempotencyKey, "delete_announcement", requestHash, resp); err != nil {
			s.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	s.logger.InfoContext(ctx, "announcement deleted", slog.String("announcement_id", req.AnnouncementId))
	return resp, nil
}
