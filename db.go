// Copyright (c) 2023 BVK Chaitanya

package kv

import "context"

type Reader interface {
	Getter
	Ranger
	Scanner
}

type Writer interface {
	Setter
	Deleter
}

type ReadWriter interface {
	Reader
	Writer
}

type Snapshot interface {
	Reader

	Discard(ctx context.Context) error
}

// Transaction represents a read-write transaction.
type Transaction interface {
	ReadWriter

	// Rollback cancels a transaction without checking for conflicts. Returns nil
	// on success.
	//
	// Rollback may return os.ErrClosed if transaction is already committed or
	// rolled-back.
	Rollback(ctx context.Context) error

	// Commit validates all reads (and writes) performed by the transaction for
	// conflicts with other transactions and atomically applies all changes to
	// the backing key-value store. Returns nil if transaction is committed
	// successfully.
	//
	// Commit may return os.ErrClosed if transaction is already committed or
	// rolled-back.
	Commit(ctx context.Context) error
}

type Database interface {
	NewTransaction(ctx context.Context) (Transaction, error)
	NewSnapshot(ctx context.Context) (Snapshot, error)
}
