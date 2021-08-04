package kv

import "context"

type Deleter interface {
	// Delete removes a key-value pair from the key-value store. Returns
	// os.ErrNotExist if no matching key exists in the key-value store.
	Delete(ctx context.Context, key string) error
}
