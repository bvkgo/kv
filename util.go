package kv

import "context"

// Type aliases for commonly required functions.
type (
	NewTxFunc   = func(context.Context) (Transaction, error)
	NewSnapFunc = func(context.Context) (Snapshot, error)
	NewIterFunc = func(context.Context) (Iterator, error)
)
