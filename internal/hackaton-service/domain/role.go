package domain

import participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"

type HackathonRole string

const (
	RoleOwner     HackathonRole = "owner"
	RoleOrganizer HackathonRole = "organizer"
	RoleMentor    HackathonRole = "mentor"
	RoleJury      HackathonRole = "jury"
)

type ParticipationStatus string

const (
	ParticipationNone           ParticipationStatus = "none"
	ParticipationIndividual     ParticipationStatus = "individual"
	ParticipationLookingForTeam ParticipationStatus = "looking_for_team"
	ParticipationTeamMember     ParticipationStatus = "team_member"
	ParticipationTeamCaptain    ParticipationStatus = "team_captain"
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
		return RoleJury
	default:
		return ""
	}
}

func MapProtoParticipationToDomain(protoStatus participationrolesv1.ParticipationStatus) ParticipationStatus {
	switch protoStatus {
	case participationrolesv1.ParticipationStatus_PART_NONE:
		return ParticipationNone
	case participationrolesv1.ParticipationStatus_PART_INDIVIDUAL:
		return ParticipationIndividual
	case participationrolesv1.ParticipationStatus_PART_LOOKING_FOR_TEAM:
		return ParticipationLookingForTeam
	case participationrolesv1.ParticipationStatus_PART_TEAM_MEMBER:
		return ParticipationTeamMember
	case participationrolesv1.ParticipationStatus_PART_TEAM_CAPTAIN:
		return ParticipationTeamCaptain
	default:
		return ParticipationNone
	}
}
