package kv

import "context"

type Iterator interface {
	// GetNext returns a key-value pair and advances the internal cursor to the
	// next key-value pair. Repeated calls to this will return all key-value
	// pairs under iteration. Returns os.ErrNotExist when all key-value pairs are
	// returned to the caller.
	GetNext(ctx context.Context) (string, string, error)
}

type Scanner interface {
	// Scan returns all key-value pairs in the database through an iterator.
	// Returned keys are not required to be in any specific order.
	Scan(ctx context.Context, it Iterator) error

	// Ascend returns key-value pairs of a selected range through an iterator, in
	// the ascending order. Every key-value pair returned is considered as READ
	// by the transaction.
	//
	// When both i and j are non-empty, the range begins with min(i,j) which is
	// included and ends at max(i,j) which is excluded.
	//
	// When one of i or j is an empty string, then range extends to the largest
	// key (inclusive). When both i and j are empty strings, then range is all
	// keys (inclusive) in the database.
	Ascend(ctx context.Context, i, j string, it Iterator) error

	// Descend is similar to Ascend, but works in the descending order.
	//
	// When both i and j are non-empty, the range begins with max(i,j) which is
	// included and ends at min(i,j) which is excluded.
	//
	// When one of i or j is an empty string, then range extends to the smallest
	// key (inclusive) in the database. When both i and j are empty strings, then
	// range is all keys (inclusive) in the database.
	Descend(ctx context.Context, i, j string, it Iterator) error
}
