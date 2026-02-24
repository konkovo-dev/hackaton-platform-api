package centrifugo

import (
	"go.uber.org/fx"
)

var Module = fx.Module("centrifugo",
	fx.Provide(
		MustNewConfig,
		NewClient,
		func(cfg *Config) *JWTHelper {
			return NewJWTHelper(cfg.JWTSecret, cfg.JWTTTL)
		},
	),
)
