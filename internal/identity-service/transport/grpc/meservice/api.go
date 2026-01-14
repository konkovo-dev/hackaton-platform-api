package meservice

import (
	"context"
	"errors"
	"log/slog"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/mappers"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/me"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
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

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		s.logger.WarnContext(ctx, "create_me: invalid user_id", slog.String("user_id", req.UserId))
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	result, err := s.meService.CreateMe(ctx, me.CreateMeIn{
		UserID:    userID,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Timezone:  req.Timezone,
	})

	if err != nil {
		return nil, s.handleError(ctx, err, "create_me")
	}

	resp := mappers.CreateMeOutToResponse(result)

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
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		s.logger.WarnContext(ctx, "get_me: no user_id in context")
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.WarnContext(ctx, "get_me: invalid user_id", slog.String("user_id", userID))
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	result, err := s.meService.GetMe(ctx, me.GetMeIn{
		UserID: userUUID,
	})
	if err != nil {
		return nil, s.handleError(ctx, err, "get_me")
	}

	s.logger.InfoContext(ctx, "get_me: success", slog.String("user_id", userID))
	return mappers.GetMeOutToResponse(result), nil
}

func (s *MeService) UpdateMe(ctx context.Context, req *identityv1.UpdateMeRequest) (*identityv1.UpdateMeResponse, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		s.logger.WarnContext(ctx, "update_me: no user_id in context")
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.WarnContext(ctx, "update_me: invalid user_id", slog.String("user_id", userID))
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	result, err := s.meService.UpdateMe(ctx, me.UpdateMeIn{
		UserID:    userUUID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		AvatarURL: req.AvatarUrl,
		Timezone:  req.Timezone,
	})
	if err != nil {
		return nil, s.handleError(ctx, err, "update_me")
	}

	s.logger.InfoContext(ctx, "update_me: success", slog.String("user_id", userID))
	return mappers.UpdateMeOutToResponse(result), nil
}

func (s *MeService) UpdateMySkills(ctx context.Context, req *identityv1.UpdateMySkillsRequest) (*identityv1.UpdateMySkillsResponse, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		s.logger.WarnContext(ctx, "update_my_skills: no user_id in context")
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.WarnContext(ctx, "update_my_skills: invalid user_id", slog.String("user_id", userID))
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	catalogSkillIDs := make([]uuid.UUID, 0, len(req.CatalogSkillIds))
	for _, idStr := range req.CatalogSkillIds {
		id, err := uuid.Parse(idStr)
		if err != nil {
			s.logger.WarnContext(ctx, "update_my_skills: invalid catalog_skill_id", slog.String("id", idStr))
			return nil, status.Error(codes.InvalidArgument, "invalid catalog_skill_id format")
		}
		catalogSkillIDs = append(catalogSkillIDs, id)
	}

	result, err := s.meService.UpdateMySkills(ctx, me.UpdateMySkillsIn{
		UserID:           userUUID,
		CatalogSkillIDs:  catalogSkillIDs,
		CustomSkills:     req.UserSkills,
		SkillsVisibility: mappers.ProtoVisibilityLevelToDomain(req.SkillsVisibility),
	})
	if err != nil {
		return nil, s.handleError(ctx, err, "update_my_skills")
	}

	s.logger.InfoContext(ctx, "update_my_skills: success", slog.String("user_id", userID))
	return mappers.UpdateMySkillsOutToResponse(result), nil
}

func (s *MeService) UpdateMyContacts(ctx context.Context, req *identityv1.UpdateMyContactsRequest) (*identityv1.UpdateMyContactsResponse, error) {
	userID, ok := auth.GetUserID(ctx)
	if !ok {
		s.logger.WarnContext(ctx, "update_my_contacts: no user_id in context")
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.WarnContext(ctx, "update_my_contacts: invalid user_id", slog.String("user_id", userID))
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	contacts := make([]*me.ContactInput, 0, len(req.Contacts))
	for _, c := range req.Contacts {
		contacts = append(contacts, &me.ContactInput{
			Type:       mappers.ProtoContactTypeToDomain(c.Contact.Type),
			Value:      c.Contact.Value,
			Visibility: mappers.ProtoVisibilityLevelToDomain(c.Visibility),
		})
	}

	result, err := s.meService.UpdateMyContacts(ctx, me.UpdateMyContactsIn{
		UserID:             userUUID,
		Contacts:           contacts,
		ContactsVisibility: mappers.ProtoVisibilityLevelToDomain(req.ContactsVisibility),
	})
	if err != nil {
		return nil, s.handleError(ctx, err, "update_my_contacts")
	}

	s.logger.InfoContext(ctx, "update_my_contacts: success", slog.String("user_id", userID))
	return mappers.UpdateMyContactsOutToResponse(result), nil
}

func (s *MeService) handleError(ctx context.Context, err error, operation string) error {
	switch {
	case errors.Is(err, me.ErrUserNotFound):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, me.ErrUserAlreadyExists):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, me.ErrInvalidInput):
		s.logger.WarnContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		s.logger.ErrorContext(ctx, operation, slog.String("error", err.Error()))
		return status.Error(codes.Internal, "internal error")
	}
}
