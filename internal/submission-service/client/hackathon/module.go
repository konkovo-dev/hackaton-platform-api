package hackathon

import (
	"context"
	"log/slog"

	"go.uber.org/fx"
)

var Module = fx.Module("hackathon-client",
	fx.Provide(
		MustNewConfig,
		NewClient,
	),
	fx.Invoke(func(lc fx.Lifecycle, client *Client, logger *slog.Logger) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				logger.Info("hackathon client initialized")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info("closing hackathon client")
				return client.Close()
			},
		})
	}),
)
