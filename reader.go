package kv

import (
	"context"
)

type Reader interface {
	// Get returns the value for a given key. Returns os.ErrNotExist when no
	// matching key is found in the key-value store.
	Get(ctx context.Context, key string) (string, error)
}
