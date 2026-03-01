package domain

const (
	EventTypeUserSkillsUpdated          = "user.skills.updated"
	EventTypeParticipationRegistered    = "participation.registered"
	EventTypeParticipationUpdated       = "participation.updated"
	EventTypeParticipationStatusChanged = "participation.status_changed"
	EventTypeParticipationTeamAssigned  = "participation.team_assigned"
	EventTypeParticipationTeamRemoved   = "participation.team_removed"
	EventTypeTeamCreated                = "team.created"
	EventTypeTeamUpdated                = "team.updated"
	EventTypeTeamDeleted                = "team.deleted"
	EventTypeVacancyCreated             = "vacancy.created"
	EventTypeVacancyUpdated             = "vacancy.updated"
	EventTypeVacancySlotsChanged        = "vacancy.slots_changed"
)

const (
	ParticipationStatusNone            = "none"
	ParticipationStatusIndividual      = "individual"
	ParticipationStatusLookingForTeam  = "looking_for_team"
	ParticipationStatusTeamMember      = "team_member"
	ParticipationStatusTeamCaptain     = "team_captain"
)

const (
	ScoreWeightSkills = 0.63
	ScoreWeightRoles  = 0.27
	ScoreWeightText   = 0.10
)

const (
	HackathonStageDraft        = "draft"
	HackathonStageUpcoming     = "upcoming"
	HackathonStageRegistration = "registration"
	HackathonStagePreStart     = "prestart"
	HackathonStageRunning      = "running"
	HackathonStageJudging      = "judging"
	HackathonStageFinished     = "finished"
)
