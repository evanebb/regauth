package store

import "context"

type TransactionStore interface {
	Tx(ctx context.Context, fn func(ctx context.Context) error) error
}
