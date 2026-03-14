package submission

import (
	"context"
	"log/slog"

	"go.uber.org/fx"
)

var Module = fx.Module("submission-client",
	fx.Provide(
		MustNewConfig,
		NewClient,
	),
	fx.Invoke(func(lc fx.Lifecycle, client *Client, logger *slog.Logger) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				logger.Info("submission client initialized")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info("closing submission client")
				return client.Close()
			},
		})
	}),
)
