package mappers

import (
	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
)

func UserToProto(user *entity.User) *identityv1.User {
	return &identityv1.User{
		UserId:    user.ID.String(),
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarUrl: user.AvatarURL,
		Timezone:  user.Timezone,
	}
}
