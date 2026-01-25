package policy

import "github.com/belikoooova/hackaton-platform-api/pkg/policy"

const (
	ActionCreateHackathon   policy.Action = "hackathon.create"
	ActionReadHackathon     policy.Action = "hackathon.read"
	ActionUpdateHackathon   policy.Action = "hackathon.update"
	ActionPublishHackathon  policy.Action = "hackathon.publish"
	ActionValidateHackathon policy.Action = "hackathon.validate"

	ActionReadTask   policy.Action = "hackathon.task.read"
	ActionUpdateTask policy.Action = "hackathon.task.update"

	ActionReadResult        policy.Action = "hackathon.result.read"
	ActionUpdateResultDraft policy.Action = "hackathon.result.update_draft"
	ActionPublishResult     policy.Action = "hackathon.result.publish"

	ActionCreateAnnouncement policy.Action = "hackathon.announcement.create"
	ActionReadAnnouncements  policy.Action = "hackathon.announcement.read"
	ActionUpdateAnnouncement policy.Action = "hackathon.announcement.update"
	ActionDeleteAnnouncement policy.Action = "hackathon.announcement.delete"
)
