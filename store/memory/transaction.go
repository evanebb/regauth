package memory

import "context"

type TransactionStore struct{}

func (t TransactionStore) Tx(ctx context.Context, fn func(ctx context.Context) error) error {
	// just pass this through, this should only be used in testing and we don't care about transactions there
	return fn(ctx)
}
