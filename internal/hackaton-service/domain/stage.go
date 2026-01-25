package domain

type HackathonStage string

const (
	StageDraft        HackathonStage = "draft"
	StageUpcoming     HackathonStage = "upcoming"
	StageRegistration HackathonStage = "registration"
	StagePreStart     HackathonStage = "prestart"
	StageRunning      HackathonStage = "running"
	StageJudging      HackathonStage = "judging"
	StageFinished     HackathonStage = "finished"
)

type HackathonState string

const (
	StateDraft     HackathonState = "draft"
	StatePublished HackathonState = "published"
)
