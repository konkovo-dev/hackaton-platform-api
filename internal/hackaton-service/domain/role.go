package domain

import participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"

type HackathonRole string

const (
	RoleOwner     HackathonRole = "owner"
	RoleOrganizer HackathonRole = "organizer"
	RoleMentor    HackathonRole = "mentor"
	RoleJudge     HackathonRole = "judge"
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
		return RoleJudge
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

func MapDomainRoleToProto(domainRole HackathonRole) participationrolesv1.HackathonRole {
	switch domainRole {
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

func MapDomainParticipationToProto(domainStatus ParticipationStatus) participationrolesv1.ParticipationStatus {
	switch domainStatus {
	case ParticipationNone:
		return participationrolesv1.ParticipationStatus_PART_NONE
	case ParticipationIndividual:
		return participationrolesv1.ParticipationStatus_PART_INDIVIDUAL
	case ParticipationLookingForTeam:
		return participationrolesv1.ParticipationStatus_PART_LOOKING_FOR_TEAM
	case ParticipationTeamMember:
		return participationrolesv1.ParticipationStatus_PART_TEAM_MEMBER
	case ParticipationTeamCaptain:
		return participationrolesv1.ParticipationStatus_PART_TEAM_CAPTAIN
	default:
		return participationrolesv1.ParticipationStatus_PARTICIPATION_STATUS_UNSPECIFIED
	}
}
