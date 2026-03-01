package outbox

import (
	pkgoutbox "github.com/belikoooova/hackaton-platform-api/pkg/outbox"
	"go.uber.org/fx"
)

var Module = fx.Module("outbox-handlers",
	fx.Provide(
		fx.Annotate(
			NewTeamCreatedHandler,
			fx.As(new(pkgoutbox.Handler)),
			fx.ResultTags(`group:"outbox_handlers"`),
		),
		fx.Annotate(
			NewTeamUpdatedHandler,
			fx.As(new(pkgoutbox.Handler)),
			fx.ResultTags(`group:"outbox_handlers"`),
		),
		fx.Annotate(
			NewTeamDeletedHandler,
			fx.As(new(pkgoutbox.Handler)),
			fx.ResultTags(`group:"outbox_handlers"`),
		),
		fx.Annotate(
			NewVacancyCreatedHandler,
			fx.As(new(pkgoutbox.Handler)),
			fx.ResultTags(`group:"outbox_handlers"`),
		),
		fx.Annotate(
			NewVacancyUpdatedHandler,
			fx.As(new(pkgoutbox.Handler)),
			fx.ResultTags(`group:"outbox_handlers"`),
		),
		fx.Annotate(
			NewVacancySlotsChangedHandler,
			fx.As(new(pkgoutbox.Handler)),
			fx.ResultTags(`group:"outbox_handlers"`),
		),
	),
)
