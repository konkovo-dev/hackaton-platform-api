package teammember

import (
	"go.uber.org/fx"
)

var Module = fx.Module("teammember-usecase",
	fx.Provide(NewService),
)
