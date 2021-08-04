package kv

import "context"

type Transaction interface {
	// Commit atomically makes all transaction changes durable and visible on the
	// backing kv store.  Returns sql.ErrTxDone if transaction was already
	// committed or rolled back.
	Commit(ctx context.Context) error

	// Rollback drops all changes performed by the transaction.  Returns
	// sql.ErrTxDone if transaction was already committed or rolled back.
	Rollback(ctx context.Context) error
}
