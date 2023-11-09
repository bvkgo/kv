// Copyright (c) 2023 BVK Chaitanya

package kvtests

import (
	"context"
	"errors"
	"io"

	"github.com/bvkgo/kv"
)

// ClearDatabase deletes all key-value pairs in the database.
func Clear(ctx context.Context, db kv.Database) error {
	clear := func(ctx context.Context, rw kv.ReadWriter) error {
		it, err := rw.Scan(ctx)
		if err != nil {
			return err
		}
		defer kv.Close(it)

		for k, _, err := it.Fetch(ctx, false); err == nil; k, _, err = it.Fetch(ctx, true) {
			if err := rw.Delete(ctx, k); err != nil {
				return err
			}
		}

		if _, _, err := it.Fetch(ctx, false); err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		return nil
	}
	return kv.WithReadWriter(ctx, db, clear)
}
