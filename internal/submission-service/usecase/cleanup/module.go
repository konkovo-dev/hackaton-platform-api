package cleanup

import (
	"context"

	"go.uber.org/fx"
)

var Module = fx.Module("cleanup",
	fx.Provide(NewProcessor),
	fx.Invoke(func(lc fx.Lifecycle, processor *Processor) {
		ctx, cancel := context.WithCancel(context.Background())

		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				go processor.Run(ctx)
				return nil
			},
			OnStop: func(_ context.Context) error {
				cancel()
				return nil
			},
		})
	}),
)
