package participationroles

import (
	"context"

	"go.uber.org/fx"
)

var Module = fx.Module("participation-roles-client",
	fx.Provide(
		MustNewConfig,
		NewClient,
	),
	fx.Invoke(func(lc fx.Lifecycle, client *Client) {
		lc.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				return client.Close()
			},
		})
	}),
)
