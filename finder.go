package kv

import "context"

// Finder functions can be used to locate a matching or the closest key-value
// pair in the key-value store. Returns os.ErrNotExist when no closest
// key-value pair is found in the key-value store.
type Finder interface {
	FindLE(ctx context.Context, key string) (string, string, error)
	FindLT(ctx context.Context, key string) (string, string, error)
	FindGE(ctx context.Context, key string) (string, string, error)
	FindGT(ctx context.Context, key string) (string, string, error)
}
