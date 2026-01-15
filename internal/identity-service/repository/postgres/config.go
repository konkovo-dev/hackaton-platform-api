package postgres

import (
	"github.com/belikoooova/hackaton-platform-api/pkg/pgxutil"
)

func NewConfig() (*pgxutil.Config, error) {
	return pgxutil.NewConfig(pgxutil.ConfigOptions{
		DefaultDSN:  "postgres://postgres:postgres@localhost:5432/hackathon?sslmode=disable",
		DSNRequired: false,
	})
}

func MustNewConfig() *pgxutil.Config {
	return pgxutil.MustNewConfig(pgxutil.ConfigOptions{
		DefaultDSN:  "postgres://postgres:postgres@localhost:5432/hackathon?sslmode=disable",
		DSNRequired: false,
	})
}
