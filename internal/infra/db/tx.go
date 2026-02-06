package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
withTx reusable wrapper function to support atomic transactional
*/
func withTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			// Re-panic so the application knows something went wrong
			panic(r)
		}
	}()

	err = fn(tx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
