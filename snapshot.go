package kv

import "context"

// Snapshot lists the methods for read-only snapshots.
type Snapshot interface {
	Reader
	Scanner

	// Discard releases a snapshot. Future operations on the snapshot may fail
	// with non-nil errors.
	Discard(ctx context.Context) error
}
