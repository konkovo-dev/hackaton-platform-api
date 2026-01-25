package policy

import "github.com/belikoooova/hackaton-platform-api/pkg/policy"

const (
	ActionCreateHackathon   policy.Action = "hackathon.create"
	ActionReadHackathon     policy.Action = "hackathon.read"
	ActionUpdateHackathon   policy.Action = "hackathon.update"
	ActionPublishHackathon  policy.Action = "hackathon.publish"
	ActionValidateHackathon policy.Action = "hackathon.validate"

	ActionCreateAnnouncement policy.Action = "hackathon.announcement.create"
	ActionReadAnnouncements  policy.Action = "hackathon.announcement.read"
	ActionUpdateAnnouncement policy.Action = "hackathon.announcement.update"
	ActionDeleteAnnouncement policy.Action = "hackathon.announcement.delete"
)
