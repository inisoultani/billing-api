package service

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
withTx reusable wrapper function to support atomic transactional
*/
func withTx[T any](ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) (T, error)) (T, error) {
	var null T
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return null, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	result, err := fn(tx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return null, err
	}

	if err := tx.Commit(ctx); err != nil {
		return null, err
	}

	return result, nil
}
