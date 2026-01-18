package domain

type HackathonStage string

const (
	StageUpcoming     HackathonStage = "hackathon_stage_upcoming"
	StageRegistration HackathonStage = "hackathon_stage_registration"
	StagePreStart     HackathonStage = "hackathon_stage_pre_start"
	StageRunning      HackathonStage = "hackathon_stage_running"
	StageJudging      HackathonStage = "hackathon_stage_judging"
	StageFinished     HackathonStage = "hackathon_stage_finished"
)
