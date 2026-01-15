package pgxutil

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UnitOfWork[T any] interface {
	Do(ctx context.Context, fn func(ctx context.Context, txRepos T) error) error
}

type unitOfWork[T any] struct {
	pool    *pgxpool.Pool
	factory func(tx pgx.Tx) T
}

func NewUnitOfWork[T any](pool *pgxpool.Pool, factory func(tx pgx.Tx) T) UnitOfWork[T] {
	return &unitOfWork[T]{
		pool:    pool,
		factory: factory,
	}
}

func (u *unitOfWork[T]) Do(ctx context.Context, fn func(ctx context.Context, txRepos T) error) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	txRepos := u.factory(tx)
	if err = fn(ctx, txRepos); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx failed: %v, rollback failed: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
