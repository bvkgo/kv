package kv

import "context"

type Writer interface {
	// Set adds a new key-value pair or updates an existing key-value pair in the
	// key-value store.
	Set(ctx context.Context, key, value string) error
}
