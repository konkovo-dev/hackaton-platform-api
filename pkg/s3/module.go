package s3

import (
	"context"
	"fmt"
	"log/slog"

	"go.uber.org/fx"
)

var Module = fx.Module("s3",
	fx.Provide(
		MustNewConfig,
		NewClient,
	),
	fx.Invoke(func(lc fx.Lifecycle, client *Client, config *Config, logger *slog.Logger) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				logger.Info("ensuring S3 bucket exists", "bucket", config.BucketName)
				if err := client.EnsureBucket(ctx, config.BucketName); err != nil {
					return fmt.Errorf("failed to ensure bucket: %w", err)
				}
				logger.Info("S3 client initialized", "endpoint", config.Endpoint, "bucket", config.BucketName)
				return nil
			},
		})
	}),
)
