package identity

import (
	"context"
	"log/slog"

	"go.uber.org/fx"
)

var Module = fx.Module("identity-client",
	fx.Provide(
		MustNewConfig,
		func(lc fx.Lifecycle, cfg *Config, logger *slog.Logger) (*Client, error) {
			client, err := NewClient(cfg)
			if err != nil {
				return nil, err
			}

			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("identity client initialized", slog.String("url", cfg.IdentityServiceURL))
					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("closing identity client connection")
					return client.Close()
				},
			})

			return client, nil
		},
	),
)
