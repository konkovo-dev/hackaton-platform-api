package pgxutil

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type BaseRepository[Q any, DB any] struct {
	queries Q
	db      DB
}

func NewBaseRepository[Q any, DB any](db DB, newQueries func(DB) Q) *BaseRepository[Q, DB] {
	return &BaseRepository[Q, DB]{
		queries: newQueries(db),
		db:      db,
	}
}

func (r *BaseRepository[Q, DB]) Queries() Q {
	return r.queries
}

func (r *BaseRepository[Q, DB]) DB() DB {
	return r.db
}
