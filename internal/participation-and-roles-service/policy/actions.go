package policy

import "github.com/belikoooova/hackaton-platform-api/pkg/policy"

const (
	ActionReadStaff         policy.Action = "staff.read"
	ActionCreateInvitation  policy.Action = "staff.invite"
	ActionCancelInvitation  policy.Action = "staff.cancel_invite"
	ActionListMyInvitations policy.Action = "staff.list_my_invitations"
	ActionAcceptInvitation  policy.Action = "staff.accept_invite"
	ActionRejectInvitation  policy.Action = "staff.reject_invite"
	ActionRemoveRole        policy.Action = "staff.remove_role"
	ActionSelfRemoveRole    policy.Action = "staff.self_remove_role"
)
