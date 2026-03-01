package outbox

import (
	pkgoutbox "github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"go.uber.org/fx"
)

var Module = fx.Module("outbox-handlers",
	fx.Provide(
		fx.Annotate(
			NewParticipationRegisteredHandler,
			fx.As(new(pkgoutbox.Handler)),
			fx.ResultTags(`group:"outbox_handlers"`),
		),
		fx.Annotate(
			NewParticipationUpdatedHandler,
			fx.As(new(pkgoutbox.Handler)),
			fx.ResultTags(`group:"outbox_handlers"`),
		),
		fx.Annotate(
			NewParticipationStatusChangedHandler,
			fx.As(new(pkgoutbox.Handler)),
			fx.ResultTags(`group:"outbox_handlers"`),
		),
		fx.Annotate(
			NewParticipationTeamAssignedHandler,
			fx.As(new(pkgoutbox.Handler)),
			fx.ResultTags(`group:"outbox_handlers"`),
		),
		fx.Annotate(
			NewParticipationTeamRemovedHandler,
			fx.As(new(pkgoutbox.Handler)),
			fx.ResultTags(`group:"outbox_handlers"`),
		),
	),
)
