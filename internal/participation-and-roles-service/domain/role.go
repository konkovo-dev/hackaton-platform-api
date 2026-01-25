package domain

import participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"

type HackathonRole string

const (
	RoleOwner     HackathonRole = "owner"
	RoleOrganizer HackathonRole = "organizer"
	RoleMentor    HackathonRole = "mentor"
	RoleJudge     HackathonRole = "judge"
)

func MapProtoRoleToDomain(protoRole participationrolesv1.HackathonRole) HackathonRole {
	switch protoRole {
	case participationrolesv1.HackathonRole_HX_ROLE_OWNER:
		return RoleOwner
	case participationrolesv1.HackathonRole_HX_ROLE_ORGANIZER:
		return RoleOrganizer
	case participationrolesv1.HackathonRole_HX_ROLE_MENTOR:
		return RoleMentor
	case participationrolesv1.HackathonRole_HX_ROLE_JUDGE:
		return RoleJudge
	default:
		return ""
	}
}

func MapDomainRoleToProto(role HackathonRole) participationrolesv1.HackathonRole {
	switch role {
	case RoleOwner:
		return participationrolesv1.HackathonRole_HX_ROLE_OWNER
	case RoleOrganizer:
		return participationrolesv1.HackathonRole_HX_ROLE_ORGANIZER
	case RoleMentor:
		return participationrolesv1.HackathonRole_HX_ROLE_MENTOR
	case RoleJudge:
		return participationrolesv1.HackathonRole_HX_ROLE_JUDGE
	default:
		return participationrolesv1.HackathonRole_HACKATHON_ROLE_UNSPECIFIED
	}
}
