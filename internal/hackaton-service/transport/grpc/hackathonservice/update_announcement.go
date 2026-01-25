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

func (s *HackathonService) UpdateHackathonAnnouncement(ctx context.Context, req *hackathonv1.UpdateHackathonAnnouncementRequest) (*hackathonv1.UpdateHackathonAnnouncementResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.AnnouncementId, req.Title, req.Body)
		resp := &hackathonv1.UpdateHackathonAnnouncementResponse{}
		found, err := s.idempotency.CheckAndGet(ctx, idempotencyKey, "update_announcement", requestHash, resp)
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

	_, err = s.announcementService.UpdateAnnouncement(ctx, announcement.UpdateAnnouncementIn{
		HackathonID:    hackathonID,
		AnnouncementID: announcementID,
		Title:          req.Title,
		Body:           req.Body,
	})
	if err != nil {
		return nil, s.handleAnnouncementError(ctx, err, "update_announcement")
	}

	resp := &hackathonv1.UpdateHackathonAnnouncementResponse{}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.HackathonId, req.AnnouncementId, req.Title, req.Body)
		if err := s.idempotency.Save(ctx, idempotencyKey, "update_announcement", requestHash, resp); err != nil {
			s.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	s.logger.InfoContext(ctx, "announcement updated", slog.String("announcement_id", req.AnnouncementId))
	return resp, nil
}
