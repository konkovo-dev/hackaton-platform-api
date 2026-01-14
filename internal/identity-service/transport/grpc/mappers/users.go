package mappers

import (
	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/users"
)

func PublicContactsToProto(contacts []*entity.Contact) []*identityv1.Contact {
	result := make([]*identityv1.Contact, 0, len(contacts))

	for _, c := range contacts {
		result = append(result, &identityv1.Contact{
			Id:    c.ID.String(),
			Type:  ParseContactType(c.Type),
			Value: c.Value,
		})
	}

	return result
}

func GetUserOutToResponse(out *users.GetUserOut) *identityv1.GetUserResponse {
	return &identityv1.GetUserResponse{
		User:     UserToProto(out.User),
		Skills:   CatalogSkillsToProto(out.Skills.Catalog, out.Skills.Custom),
		Contacts: PublicContactsToProto(out.Contacts),
	}
}

func BatchGetUsersOutToResponse(out *users.BatchGetUsersOut) *identityv1.BatchGetUsersResponse {
	usersResponses := make([]*identityv1.GetUserResponse, 0, len(out.Users))

	for _, user := range out.Users {
		usersResponses = append(usersResponses, GetUserOutToResponse(user))
	}

	return &identityv1.BatchGetUsersResponse{
		Users: usersResponses,
	}
}

func ListUsersOutToResponse(out *users.ListUsersOut) *identityv1.ListUsersResponse {
	usersResponses := make([]*identityv1.GetUserResponse, 0, len(out.Users))

	for _, user := range out.Users {
		usersResponses = append(usersResponses, GetUserOutToResponse(user))
	}

	return &identityv1.ListUsersResponse{
		Users: usersResponses,
		Page: &commonv1.PageResponse{
			NextPageToken: out.NextPageToken,
		},
	}
}
