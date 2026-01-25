package domain

type HackathonStage string

const (
	StageDraft        HackathonStage = "hackathon_stage_draft"
	StageUpcoming     HackathonStage = "hackathon_stage_upcoming"
	StageRegistration HackathonStage = "hackathon_stage_registration"
	StagePreStart     HackathonStage = "hackathon_stage_pre_start"
	StageRunning      HackathonStage = "hackathon_stage_running"
	StageJudging      HackathonStage = "hackathon_stage_judging"
	StageFinished     HackathonStage = "hackathon_stage_finished"
)

type HackathonState string

const (
	StateDraft     HackathonState = "hackathon_state_draft"
	StatePublished HackathonState = "hackathon_state_published"
)
