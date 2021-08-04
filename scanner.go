package kv

import "context"

type Scanner interface {
	// Scan runs the user callback for every key-value pair in the key-value
	// store. Callback may be run in random order of the keys.
	//
	// When called from transactions, every key-value pair passed to the callback
	// is considered as READ, which may impact the transaction commit behavior.
	Scan(ctx context.Context, cb func(context.Context, string, string) error) error
}
