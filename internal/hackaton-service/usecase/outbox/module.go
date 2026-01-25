package outbox

import (
	pkgoutbox "github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"go.uber.org/fx"
)

var Module = fx.Module("outbox-handlers",
	fx.Provide(
		fx.Annotate(
			NewOwnerAssignedHandler,
			fx.As(new(pkgoutbox.Handler)),
			fx.ResultTags(`group:"outbox_handlers"`),
		),
	),
)
