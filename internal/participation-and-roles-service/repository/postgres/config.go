package postgres

import (
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
)

func NewConfig() (*pgxutil.Config, error) {
	return pgxutil.NewConfig(pgxutil.ConfigOptions{
		DSNRequired: true,
	})
}

func MustNewConfig() *pgxutil.Config {
	return pgxutil.MustNewConfig(pgxutil.ConfigOptions{
		DSNRequired: true,
	})
}
