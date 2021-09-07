package kv

import "context"

// Transaction lists the methods for read/write transactions.
type Transaction interface {
	Reader
	Scanner

	Writer
	Deleter

	// Commit atomically makes all transaction changes durable and visible on the
	// backing kv store.  Returns sql.ErrTxDone if transaction was already
	// committed or rolled back.
	Commit(ctx context.Context) error

	// Discard drops all changes performed by the transaction.  Returns
	// sql.ErrTxDone if transaction was already committed or discarded.
	Discard(ctx context.Context) error
}
