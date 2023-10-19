// Copyright (c) 2023 BVK Chaitanya

package kv

import (
	"io"
)

// Close is a helper function that invokes Close method on the Iterator or
// Snapshot or a Transaction if such a method is defined; otherwise it is a
// no-op.
//
// Some implementations may require to release resources, so this helper
// function can be used close an iterator.
func Close(v any) error {
	if c, ok := v.(io.Closer); ok {
		c.Close()
		return nil
	}
	if c, ok := v.(interface{ Close() }); ok {
		c.Close()
		return nil
	}
	return nil
}
