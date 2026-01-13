package meservice

import (
	"context"
	"errors"
	"log/slog"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/me"
	"github.com/belikoooova/hackaton-platform-api/pkg/idempotency"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MeService struct {
	identityv1.UnimplementedMeServiceServer
	meService   *me.Service
	idempotency *idempotency.Helper
	logger      *slog.Logger
}

func NewMeService(
	meService *me.Service,
	idempotencyHelper *idempotency.Helper,
	logger *slog.Logger,
) *MeService {
	return &MeService{
		meService:   meService,
		idempotency: idempotencyHelper,
		logger:      logger,
	}
}

func (s *MeService) CreateMe(ctx context.Context, req *identityv1.CreateMeRequest) (*identityv1.CreateMeResponse, error) {
	var idempotencyKey string
	if req.IdempotencyKey != nil {
		idempotencyKey = req.IdempotencyKey.Key
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.UserId, req.Username)
		resp := &identityv1.CreateMeResponse{}
		found, err := s.idempotency.CheckAndGet(ctx, idempotencyKey, "create_me", requestHash, resp)
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

	if req.UserId == "" {
		s.logger.WarnContext(ctx, "create_me: empty user_id")
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		s.logger.WarnContext(ctx, "create_me: invalid user_id", slog.String("user_id", req.UserId))
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	if req.Username == "" {
		s.logger.WarnContext(ctx, "create_me: empty username")
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	if req.FirstName == "" {
		s.logger.WarnContext(ctx, "create_me: empty first_name")
		return nil, status.Error(codes.InvalidArgument, "first_name is required")
	}

	if req.LastName == "" {
		s.logger.WarnContext(ctx, "create_me: empty last_name")
		return nil, status.Error(codes.InvalidArgument, "last_name is required")
	}

	if req.Timezone == "" {
		s.logger.WarnContext(ctx, "create_me: empty timezone")
		return nil, status.Error(codes.InvalidArgument, "timezone is required")
	}

	result, err := s.meService.CreateMe(ctx, me.CreateMeInput{
		UserID:    userID,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Timezone:  req.Timezone,
	})

	if err != nil {
		return nil, s.handleError(ctx, err, "create_me")
	}

	resp := &identityv1.CreateMeResponse{
		User: &identityv1.User{
			UserId:    result.User.ID.String(),
			Username:  result.User.Username,
			FirstName: result.User.FirstName,
			LastName:  result.User.LastName,
			AvatarUrl: result.User.AvatarURL,
			Timezone:  result.User.Timezone,
		},
	}

	if idempotencyKey != "" {
		requestHash := idempotency.ComputeHash(req.UserId, req.Username)
		if err := s.idempotency.Save(ctx, idempotencyKey, "create_me", requestHash, resp); err != nil {
			s.logger.ErrorContext(ctx, "failed to save idempotency key", slog.String("error", err.Error()))
		}
	}

	s.logger.InfoContext(ctx, "user created", slog.String("user_id", result.User.ID.String()))
	return resp, nil
}

func (s *MeService) GetMe(ctx context.Context, req *identityv1.GetMeRequest) (*identityv1.GetMeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented yet")
}

func (s *MeService) UpdateMe(ctx context.Context, req *identityv1.UpdateMeRequest) (*identityv1.UpdateMeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented yet")
}

func (s *MeService) UpdateMySkills(ctx context.Context, req *identityv1.UpdateMySkillsRequest) (*identityv1.UpdateMySkillsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented yet")
}

func (s *MeService) UpdateMyContacts(ctx context.Context, req *identityv1.UpdateMyContactsRequest) (*identityv1.UpdateMyContactsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented yet")
}

func (s *MeService) handleError(ctx context.Context, err error, operation string) error {
	switch {
	case errors.Is(err, me.ErrUserNotFound):
		s.logger.WarnContext(ctx, operation+": user not found")
		return status.Error(codes.NotFound, "user not found")
	case errors.Is(err, me.ErrUserAlreadyExists):
		s.logger.WarnContext(ctx, operation+": user already exists")
		return status.Error(codes.AlreadyExists, "user already exists")
	case errors.Is(err, me.ErrInvalidInput):
		s.logger.WarnContext(ctx, operation+": invalid input")
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		s.logger.ErrorContext(ctx, operation+": internal error", slog.String("error", err.Error()))
		return status.Error(codes.Internal, "internal error")
	}
}
