package domain

// Ticket statuses
const (
	TicketStatusOpen   = "open"
	TicketStatusClosed = "closed"
)

// Author roles
const (
	AuthorRoleParticipant = "participant"
	AuthorRoleMentor      = "mentor"
	AuthorRoleOrganizer   = "organizer"
	AuthorRoleSystem      = "system"
)

// Owner kinds
const (
	OwnerKindUser = "user"
	OwnerKindTeam = "team"
)

// Hackathon stages (from hackaton-service)
const (
	HackathonStageDraft        = "draft"
	HackathonStageUpcoming     = "upcoming"
	HackathonStageRegistration = "registration"
	HackathonStagePreStart     = "prestart"
	HackathonStageRunning      = "running"
	HackathonStageJudging      = "judging"
	HackathonStageFinished     = "finished"
)

// Participation statuses
const (
	ParticipationStatusActive = "active"
)

// Hackathon roles
const (
	RoleMentor    = "mentor"
	RoleOrganizer = "organizer"
	RoleOwner     = "owner"
)
