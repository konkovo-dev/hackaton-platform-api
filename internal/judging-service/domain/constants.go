package domain

const (
	OwnerKindUser = "user"
	OwnerKindTeam = "team"
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
	MinScore = 0
	MaxScore = 10
)

const (
	MinCommentLength = 1
	MaxCommentLength = 5000
)

const (
	MinJudgesPerSubmission = 3
)
