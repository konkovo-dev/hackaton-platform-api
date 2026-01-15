package mappers

import (
	"strings"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/me"
)

func ContactsToProto(contacts []*me.GetMeContact) []*identityv1.MyContact {
	result := make([]*identityv1.MyContact, 0, len(contacts))

	for _, c := range contacts {
		result = append(result, &identityv1.MyContact{
			Contact: &identityv1.Contact{
				Id:    c.Contact.ID.String(),
				Type:  ParseContactType(c.Contact.Type),
				Value: c.Contact.Value,
			},
			Visibility: ParseVisibilityLevel(c.Visibility),
		})
	}

	return result
}

func ParseContactType(contactType string) identityv1.ContactType {
	switch strings.ToLower(contactType) {
	case "email":
		return identityv1.ContactType_CONTACT_TYPE_EMAIL
	case "telegram":
		return identityv1.ContactType_CONTACT_TYPE_TELEGRAM
	case "github":
		return identityv1.ContactType_CONTACT_TYPE_GITHUB
	case "linkedin":
		return identityv1.ContactType_CONTACT_TYPE_LINKEDIN
	default:
		return identityv1.ContactType_CONTACT_TYPE_UNSPECIFIED
	}
}

func ProtoContactTypeToDomain(contactType identityv1.ContactType) string {
	switch contactType {
	case identityv1.ContactType_CONTACT_TYPE_EMAIL:
		return "email"
	case identityv1.ContactType_CONTACT_TYPE_TELEGRAM:
		return "telegram"
	case identityv1.ContactType_CONTACT_TYPE_GITHUB:
		return "github"
	case identityv1.ContactType_CONTACT_TYPE_LINKEDIN:
		return "linkedin"
	default:
		return ""
	}
}

func EntityContactsToProto(contacts []*entity.Contact) []*identityv1.MyContact {
	result := make([]*identityv1.MyContact, 0, len(contacts))

	for _, c := range contacts {
		result = append(result, &identityv1.MyContact{
			Contact: &identityv1.Contact{
				Id:    c.ID.String(),
				Type:  ParseContactType(c.Type),
				Value: c.Value,
			},
			Visibility: ParseVisibilityLevel(string(c.Visibility)),
		})
	}

	return result
}
