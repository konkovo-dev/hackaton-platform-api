package vacancy

import (
	"go.uber.org/fx"
)

var Module = fx.Module("vacancy-usecase",
	fx.Provide(NewService),
)
