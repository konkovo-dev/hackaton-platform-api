package postgres

import (
	"github.com/belikoooova/hackaton-platform-api/pkg/pgx"
)

func NewConfig() (*pgx.Config, error) {
	return pgx.NewConfig(pgx.ConfigOptions{
		DefaultDSN:  "postgres://postgres:postgres@localhost:5432/hackathon?sslmode=disable",
		DSNRequired: false,
	})
}

func MustNewConfig() *pgx.Config {
	return pgx.MustNewConfig(pgx.ConfigOptions{
		DefaultDSN:  "postgres://postgres:postgres@localhost:5432/hackathon?sslmode=disable",
		DSNRequired: false,
	})
}
