package policy

import (
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
)

const (
	ActionListTeams              policy.Action = "list_teams"
	ActionGetTeam                policy.Action = "get_team"
	ActionCreateTeam             policy.Action = "create_team"
	ActionUpdateTeam             policy.Action = "update_team"
	ActionDeleteTeam             policy.Action = "delete_team"
	ActionUpsertVacancy          policy.Action = "upsert_vacancy"
	ActionListTeamMembers        policy.Action = "list_team_members"
	ActionListTeamInvitations    policy.Action = "list_team_invitations"
	ActionCreateTeamInvitation   policy.Action = "create_team_invitation"
	ActionCancelTeamInvitation   policy.Action = "cancel_team_invitation"
	ActionListMyTeamInvitations  policy.Action = "list_my_team_invitations"
	ActionAcceptTeamInvitation   policy.Action = "accept_team_invitation"
	ActionRejectTeamInvitation   policy.Action = "reject_team_invitation"
	ActionKickTeamMember         policy.Action = "kick_team_member"
	ActionLeaveTeam              policy.Action = "leave_team"
	ActionTransferCaptain        policy.Action = "transfer_captain"
	ActionListJoinRequests       policy.Action = "list_join_requests"
	ActionCreateJoinRequest      policy.Action = "create_join_request"
	ActionListMyJoinRequests     policy.Action = "list_my_join_requests"
	ActionCancelJoinRequest      policy.Action = "cancel_join_request"
	ActionAcceptJoinRequest      policy.Action = "accept_join_request"
	ActionRejectJoinRequest      policy.Action = "reject_join_request"
)
