package postgres

import (
	"github.com/belikoooova/hackaton-platform-api/pkg/pgx"
)

func NewConfig() (*pgx.Config, error) {
	return pgx.NewConfig(pgx.ConfigOptions{
		DSNRequired: true,
	})
}

func MustNewConfig() *pgx.Config {
	return pgx.MustNewConfig(pgx.ConfigOptions{
		DSNRequired: true,
	})
}
