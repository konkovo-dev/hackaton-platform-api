package meservice

import (
	"context"
	"log/slog"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/transport/grpc/mappers"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/me"
	"github.com/belikoooova/hackaton-platform-api/pkg/auth"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
