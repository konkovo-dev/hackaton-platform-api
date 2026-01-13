package outbox

import (
	"context"
	"log/slog"

	"go.uber.org/fx"
)

var Module = fx.Module("outbox",
	fx.Provide(
		MustNewConfig,
		NewProcessor,
	),
	fx.Invoke(
		fx.Annotate(
			func(lc fx.Lifecycle, processor *Processor, handlers []Handler, logger *slog.Logger) {
				for _, handler := range handlers {
					processor.RegisterHandler(handler)
				}

				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						go func() {
							logger.Info("outbox processor starting")
							if err := processor.Run(context.Background()); err != nil {
								logger.Error("outbox processor stopped", slog.String("error", err.Error()))
							}
						}()
						return nil
					},
				})
			},
			fx.ParamTags(``, ``, `group:"outbox_handlers"`, ``),
		),
	),
)
