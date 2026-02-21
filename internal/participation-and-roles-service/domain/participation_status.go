package domain

import participationrolesv1 "github.com/belikoooova/hackaton-platform-api/api/participationandroles/v1"

type ParticipationStatus string

const (
	ParticipationNone           ParticipationStatus = "none"
	ParticipationIndividual     ParticipationStatus = "individual"
	ParticipationLookingForTeam ParticipationStatus = "looking_for_team"
	ParticipationTeamMember     ParticipationStatus = "team_member"
	ParticipationTeamCaptain    ParticipationStatus = "team_captain"
)

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

func MapDomainParticipationToProto(status ParticipationStatus) participationrolesv1.ParticipationStatus {
	switch status {
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
