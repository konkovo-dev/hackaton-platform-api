package mappers

import (
	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/me"
)

func CreateMeOutToResponse(out *me.CreateMeOut) *identityv1.CreateMeResponse {
	return &identityv1.CreateMeResponse{
		User: UserToProto(out.User),
	}
}

func GetMeOutToResponse(out *me.GetMeOut) *identityv1.GetMeResponse {
	return &identityv1.GetMeResponse{
		User:       UserToProto(out.User),
		Skills:     CatalogSkillsToProto(out.Skills.Catalog, out.Skills.Custom),
		Contacts:   ContactsToProto(out.Contacts),
		Visibility: VisibilityToProto(out.Visibility),
	}
}

func UpdateMeOutToResponse(out *me.UpdateMeOut) *identityv1.UpdateMeResponse {
	return &identityv1.UpdateMeResponse{
		User: UserToProto(out.User),
	}
}

func UpdateMySkillsOutToResponse(out *me.UpdateMySkillsOut) *identityv1.UpdateMySkillsResponse {
	return &identityv1.UpdateMySkillsResponse{
		Skills:     CatalogSkillsToProto(out.CatalogSkills, out.CustomSkills),
		Visibility: VisibilityToProto(out.Visibility),
	}
}

func UpdateMyContactsOutToResponse(out *me.UpdateMyContactsOut) *identityv1.UpdateMyContactsResponse {
	return &identityv1.UpdateMyContactsResponse{
		Contacts:   EntityContactsToProto(out.Contacts),
		Visibility: VisibilityToProto(out.Visibility),
	}
}
