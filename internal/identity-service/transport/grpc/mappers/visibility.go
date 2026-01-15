package mappers

import (
	"strings"

	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
)

func VisibilityToProto(visibility *entity.Visibility) *identityv1.VisibilitySettings {
	return &identityv1.VisibilitySettings{
		Skills:   ParseVisibilityLevel(string(visibility.SkillsVisibility)),
		Contacts: ParseVisibilityLevel(string(visibility.ContactsVisibility)),
	}
}

func ParseVisibilityLevel(level string) identityv1.VisibilityLevel {
	switch strings.ToLower(level) {
	case "public":
		return identityv1.VisibilityLevel_VISIBILITY_LEVEL_PUBLIC
	case "private":
		return identityv1.VisibilityLevel_VISIBILITY_LEVEL_PRIVATE
	default:
		return identityv1.VisibilityLevel_VISIBILITY_LEVEL_UNSPECIFIED
	}
}

func ProtoVisibilityLevelToDomain(level identityv1.VisibilityLevel) domain.VisibilityLevel {
	switch level {
	case identityv1.VisibilityLevel_VISIBILITY_LEVEL_PUBLIC:
		return domain.VisibilityLevelPublic
	case identityv1.VisibilityLevel_VISIBILITY_LEVEL_PRIVATE:
		return domain.VisibilityLevelPrivate
	default:
		return domain.VisibilityLevelPublic
	}
}
