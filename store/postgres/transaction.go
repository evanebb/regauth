package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txCtxKey struct{}

type TransactionStore struct {
	db *pgxpool.Pool
}

func (s TransactionStore) QuerierFromContext(ctx context.Context) Querier {
	tx, ok := ctx.Value(txCtxKey{}).(pgx.Tx)
	if !ok {
		return s.db
	}

	return tx
}

func (s TransactionStore) Tx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	if err := fn(context.WithValue(ctx, txCtxKey{}, tx)); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
