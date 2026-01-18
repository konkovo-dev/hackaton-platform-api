package domain

type HackathonRole string

const (
	RoleOwner     HackathonRole = "owner"
	RoleOrganizer HackathonRole = "organizer"
	RoleMentor    HackathonRole = "mentor"
	RoleJudge     HackathonRole = "judge"
)
