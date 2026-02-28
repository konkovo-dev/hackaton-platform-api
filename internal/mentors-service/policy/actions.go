package policy

import (
	"github.com/belikoooova/hackaton-platform-api/pkg/policy"
)

const (
	ActionSendMessage          policy.Action = "send_message"
	ActionGetMyChatMessages    policy.Action = "get_my_chat_messages"
	ActionGetTicketMessages    policy.Action = "get_ticket_messages"
	ActionReplyInTicket        policy.Action = "reply_in_ticket"
	ActionCloseTicket          policy.Action = "close_ticket"
	ActionListAssignedTickets  policy.Action = "list_assigned_tickets"
	ActionListAllTickets       policy.Action = "list_all_tickets"
	ActionClaimTicket          policy.Action = "claim_ticket"
	ActionAssignTicket         policy.Action = "assign_ticket"
	ActionGetRealtimeToken     policy.Action = "get_realtime_token"
)
