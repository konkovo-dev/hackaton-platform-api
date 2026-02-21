package teaminbox

import (
	"go.uber.org/fx"
)

var Module = fx.Module("teaminbox-usecase",
	fx.Provide(NewService),
)
