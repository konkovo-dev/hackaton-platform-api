package domain

const (
	OwnerKindUser = "user"
	OwnerKindTeam = "team"
)

const (
	FileUploadStatusPending   = "pending"
	FileUploadStatusCompleted = "completed"
	FileUploadStatusFailed    = "failed"
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

const (
	RoleOwner     = "owner"
	RoleOrganizer = "organizer"
	RoleMentor    = "mentor"
	RoleJudge     = "judge"
)

const (
	ParticipationStatusActive = "active"
)
