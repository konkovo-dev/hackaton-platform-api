package outbox

import (
	"github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"go.uber.org/fx"
)

var Module = fx.Module("outbox",
	fx.Provide(
		NewMessageCreatedHandler,
		NewTicketClosedHandler,
		NewTicketAssignedHandler,
		fx.Annotate(
			func(
				messageCreatedHandler *MessageCreatedHandler,
				ticketClosedHandler *TicketClosedHandler,
				ticketAssignedHandler *TicketAssignedHandler,
			) []outbox.Handler {
				return []outbox.Handler{
					messageCreatedHandler,
					ticketClosedHandler,
					ticketAssignedHandler,
				}
			},
			fx.ResultTags(`group:"outbox_handlers"`),
		),
	),
)
